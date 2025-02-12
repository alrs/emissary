package ascache

import (
	"context"
	"time"

	"github.com/EmissarySocial/emissary/tools/cacheheader"
	"github.com/benpate/derp"
	"github.com/benpate/exp"
	"github.com/benpate/hannibal/streams"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	collection     *mongo.Collection
	innerClient    streams.Client
	purgeFrequency int64
	cacheMode      string
	obeyHeaders    bool
}

// New returns a fully initialized Client object
func New(innerClient streams.Client, collection *mongo.Collection, options ...ClientOptionFunc) *Client {

	// Create a default client
	result := Client{
		collection:     collection,
		innerClient:    innerClient,
		purgeFrequency: 60 * 60 * 4, // Default purge frequency is 4 hours
		cacheMode:      CacheModeReadWrite,
		obeyHeaders:    true,
	}

	// Apply option functions to the client
	result.WithOptions(options...)

	go result.start()

	return &result
}

func (client *Client) WithOptions(options ...ClientOptionFunc) {
	for _, option := range options {
		option(client)
	}
}

/******************************************
 * Lifecycle Methods
 ******************************************/

// start is a background process that purges expired documents from the cache
func (client *Client) start() {

	// If the client is not writable then don't purge expired documents
	if client.NotWritable() {
		return
	}

	// If the purge frequency is 0 then don't purge expired documents
	if client.purgeFrequency == 0 {
		return
	}

	for {
		// wait for the purge frequency duration
		time.Sleep(time.Duration(client.purgeFrequency) * time.Second)

		criteria := exp.LessThan("expires", time.Now().Unix())

		// Try to remove expired actors
		if _, err := client.collection.DeleteMany(context.Background(), criteria); err != nil {
			derp.Report(derp.Wrap(err, "ascache.Client.delete", "Error purging expired actors from cache"))
		}
	}
}

/******************************************
 * Hannibal HTTP Client Methods
 ******************************************/

func (client *Client) Load(uri string, options ...any) (streams.Document, error) {

	config := NewLoadConfig(options...)

	// If we're not forcing the cache to reload, then try to load from the cache first
	if client.IsReadable() && config.isCacheAllowed() {

		// Search the cache for the document
		cachedValue := NewCachedValue()
		if err := client.loadByURI(uri, &cachedValue); err == nil {

			// If we're allowed to write to the cache, then do it.
			if client.IsWritable() && cachedValue.ShouldRevalidate() {
				go client.revalidate(uri, options...)
			}

			result := client.asDocument(cachedValue)
			return result, nil
		}
	}

	// Pass the request to the inner client
	result, err := client.innerClient.Load(uri, options...)

	if err != nil {
		return result, derp.Wrap(err, "ascache.Client.LoadActor", "error loading document from inner client", uri)
	}

	result.WithOptions(streams.WithClient(client))

	// Try to save the new value asynchronously
	if client.IsWritable() {
		go client.save(uri, result)
	}

	return result, nil
}

/******************************************
 * Other Cache Management Methods
 ******************************************/

// revalidate reloads a document from the source even if it has not yet expired.
// This potentially updates the cache timeout value, keeping the document
// fresh in the cache for longer.
func (client *Client) revalidate(uri string, options ...any) {

	// If the client is not writable, then don't try to refresh the cache
	if client.NotWritable() {
		return
	}

	// Pass the request to the inner client
	log.Trace().Str("uri", uri).Msg("ascache.Client.revalidate")
	if result, err := client.innerClient.Load(uri, options...); err == nil {
		client.save(uri, result)
	}
}

func (client *Client) save(uri string, document streams.Document) {

	const location = "ascache.Client.save"

	// If the client is not writable, then don't try to save the document
	if client.NotWritable() {
		return
	}

	// Create a new cachedValue
	cachedValue := NewCachedValue()
	cachedValue.URI = uri
	cachedValue.Original = document.Map()
	cachedValue.HTTPHeader = document.HTTPHeader()
	cachedValue.HTTPHeader.Set(headerHannibalCache, "true")
	cachedValue.HTTPHeader.Set(headerHannibalCacheDate, time.Now().Format(time.RFC3339))

	if inReplyTo := document.InReplyTo(); inReplyTo.NotNil() {
		cachedValue.InReplyTo = inReplyTo.String()
	}

	// Calculate caching rules
	cacheControl := cacheheader.Parse(cachedValue.HTTPHeader)

	if client.obeyHeaders && cacheControl.NotCacheAllowed() {
		return
	}

	cachedValue.calcPublished()
	cachedValue.calcExpires(cacheControl)
	cachedValue.calcRevalidates(cacheControl)

	// Try to upsert the document into the cache
	filter := bson.M{"uri": uri}
	update := bson.M{"$set": cachedValue}
	queryOptions := options.Update().SetUpsert(true)

	// Try to remove any existing documents with the same URI
	if _, err := client.collection.UpdateOne(context.Background(), filter, update, queryOptions); err != nil {
		derp.Report(derp.Wrap(err, location, "Error saving document to cache", document.ID()))
	}

	// Write to log
	log.Trace().Str("uri", uri).Msg("ascache.Client.save")
}

// asDocument converts a CachedValue into a fully-populated streams.Document
func (client *Client) asDocument(cachedValue CachedValue) streams.Document {
	result := streams.NewDocument(
		cachedValue.Original,
		streams.WithClient(client),
		streams.WithHTTPHeader(cachedValue.HTTPHeader),
	)
	return result
}

/******************************************
 * Other Queries
 ******************************************/

// loadByURI loads a CachedValue from the cache using its URI.
func (client *Client) loadByURI(uri string, document *CachedValue) error {

	// Prevent NPE
	if client.collection == nil {
		return derp.NewInternalError("ascache.Client.loadByURI", "Cache connection is not defined")
	}

	// Query the cache database
	criteria := bson.M{"uri": uri}
	if err := client.collection.FindOne(context.Background(), criteria).Decode(document); err != nil {
		log.Trace().Str("uri", uri).Msg("ascache.Client.loadByURI: NOT FOUND")
		return derp.Wrap(err, "ascache.Client.loadByURI", "Error loading document from cache", uri)
	}

	// Success.
	log.Trace().Str("uri", uri).Msg("ascache.Client.loadByURI: FOUND")
	return nil
}

/******************************************
 * Configuration Accessors
 ******************************************/

// IsReadWritable returns TRUE if the cache can be read and written
func (client *Client) IsReadWritable() bool {
	return client.cacheMode == CacheModeReadWrite
}

// NotReadWritable returns TRUE if the cache cannot be read or written
func (client *Client) NotReadWritable() bool {
	return client.cacheMode != CacheModeReadWrite
}

// IsReadable returns TRUE if the client is configured to read from the cache
func (client *Client) IsReadable() bool {
	return client.cacheMode != CacheModeWriteOnly
}

// NotReadable returns TRUE if the client is not configured to read from the cache
func (client *Client) NotReadable() bool {
	return client.cacheMode == CacheModeWriteOnly
}

// isWritable returns TRUE if the client is configured to write to the cache
func (client *Client) IsWritable() bool {
	return client.cacheMode != CacheModeReadOnly
}

// NotWritable returns TRUE if the client is not configured to write to the cache
func (client *Client) NotWritable() bool {
	return client.cacheMode == CacheModeReadOnly
}
