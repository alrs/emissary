package service

import (
	"bytes"
	"mime"
	"net/http"
	"net/url"

	"github.com/EmissarySocial/emissary/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/benpate/derp"
	"github.com/benpate/digit"
	"github.com/benpate/remote"
	"github.com/benpate/rosetta/list"
	"github.com/davecgh/go-spew/spew"
	"github.com/mmcdole/gofeed"
	"github.com/tomnomnom/linkheader"
)

// discoverLinks attempts to discover ActivityPub/RSS/Atom/JSONFeed links from a given following URL.
func discoverLinks(response *http.Response, body *bytes.Buffer) []digit.Link {

	// Look for links embedded in the HTML
	if result := discoverLinks_HTML(response, body); len(result) > 0 {
		return result
	}

	// Fall back to WebFinger, just in case
	if result := discoverLinks_WebFinger(response.Request.URL.String()); len(result) > 0 {
		return result
	}

	// Fall through, fail through
	return make([]digit.Link, 0)
}

func discoverLinks_HTML(response *http.Response, body *bytes.Buffer) []digit.Link {

	const location = "service.discoverLinks_HTML"

	result := discoverLinks_Headers(response)

	// If the document itself is an RSS feed, then we're done.  Add it to the list.
	// TODO: LOW: Possibly parse RSS-Cloud here?
	mimeType := response.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(mimeType)

	switch mediaType {
	case
		model.MimeTypeJSONFeed,
		model.MimeTypeAtom,
		model.MimeTypeRSS,
		model.MimeTypeXML,
		model.MimeTypeXMLText:

		return append(result, digit.Link{
			RelationType: model.LinkRelationSelf,
			MediaType:    mediaType,
			Href:         response.Request.URL.String(),
		})
	}

	// Fall through assumes that this is an HTML document.
	// So, look for embedded links to other feeds (ActivityPub/RSS/Atom/JSONFeed).

	// Scan the HTML document for relevant links
	htmlDocument, err := goquery.NewDocumentFromReader(bytes.NewReader(body.Bytes()))

	if err != nil {
		derp.Report(derp.Wrap(err, location, "Error parsing HTML document"))
		return result
	}

	links := htmlDocument.Find("[rel=alternate],[rel=self],[rel=hub]").Nodes

	// Look through RSS links for all valid feeds
	for _, link := range links {

		relationType := nodeAttribute(link, "rel")
		href := nodeAttribute(link, "href")
		href = getRelativeURL(response.Request.URL.String(), href)

		// Special case for WebSub relation types
		switch relationType {
		case model.LinkRelationHub:

			result = append(result, digit.Link{
				RelationType: model.LinkRelationHub,
				MediaType:    model.MagicMimeTypeWebSub,
				Href:         href,
			})
			continue
		}

		// General case for all other relation types
		mediaType := nodeAttribute(link, "type")
		mediaType = list.Semicolon(mediaType).First()

		switch mediaType {

		case
			model.MimeTypeActivityPub,
			model.MimeTypeJSONFeed,
			model.MimeTypeAtom,
			model.MimeTypeRSS,
			model.MimeTypeXML,
			model.MimeTypeXMLText:

			result = append(result, digit.Link{
				RelationType: model.LinkRelationAlternate,
				MediaType:    mediaType,
				Href:         href,
			})
		}
	}

	return result
}

// discoverLinks_Headers scans the HTTP headers for WebSub links
func discoverLinks_Headers(response *http.Response) []digit.Link {

	result := make([]digit.Link, 0)

	// Scan the response headers for WebSub links
	// TODO: LOW: Are RSS links ever put into the headers?
	// TODO: LOW: Are RSSCloud links ever put into the headers?
	linkHeaders := linkheader.ParseMultiple(response.Header["Link"])

	for _, link := range linkHeaders {

		switch link.Rel {
		case model.LinkRelationHub:
			result = append(result, digit.Link{
				MediaType:    model.MagicMimeTypeWebSub,
				RelationType: link.Rel,
				Href:         link.URL,
			})

		case model.LinkRelationSelf:
			result = append(result, digit.Link{
				RelationType: link.Rel,
				Href:         link.URL,
			})

		}
	}

	return result
}

func discoverLinks_RSS(response *http.Response, rssFeed *gofeed.Feed) []digit.Link {

	result := discoverLinks_Headers(response)

	// Look for WebSub links
	for _, link := range rssFeed.Links {
		spew.Dump("found link in RSS/Atom Feed", link)
	}

	return result
}

// discoverLinks_WebFinger uses the WebFinger protocol to search for additional metadata about the targetURL
func discoverLinks_WebFinger(targetURL string) []digit.Link {

	// Compute the WebFinger service for the targetURL
	webfingerURL, err := getWebFingerURL(targetURL)

	if err != nil {
		return make([]digit.Link, 0)
	}

	// Send a GET request to the WebFinger service
	object := digit.NewResource("")
	transaction := remote.Get(webfingerURL.String()).Response(&object, nil)

	if err := transaction.Send(); err != nil {
		return make([]digit.Link, 0)
	}

	if object.Links == nil {
		return make([]digit.Link, 0)
	}

	return object.Links
}

// getWebFingerURL determines the best WebFinger URL for a given target URL.
func getWebFingerURL(targetURL string) (url.URL, error) {

	const location = "service.getWebFingerURL"
	var result url.URL

	// Try to parse the followingURL as a standard URL
	if parsedURL, err := url.Parse(targetURL); err == nil {

		result.Scheme = parsedURL.Scheme
		result.Host = parsedURL.Host
		result.Path = "/.well-known/webfinger"
		result.RawQuery = "resource=" + targetURL

		return result, nil
	}

	// TODO: HIGH: Try to parse as a Mastodon username @benpate@mastodon.social => https://mastodon.social/.well-known/webfinger?resource=acct:benpate
	// TODO: MEDIUM: Try to parse as an email address ben@pate.org => https://pate.org/.well-known/webfinger?resource=acct:ben@pate.org
	// TODO: LOW: Look into Textcasting? http://textcasting.org

	return result, derp.NewNotFoundError(location, "Error parsing following URL", targetURL)
}
