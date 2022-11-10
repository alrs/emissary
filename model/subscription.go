package model

import (
	"time"

	"github.com/benpate/data/journal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionMethodRSS represents an RSS subscription
const SubscriptionMethodRSS = "RSS"

// SubscriptionMethodWebSub represents a WebSub subscription
const SubscriptionMethodWebSub = "WEBSUB"

type Subscription struct {
	SubscriptionID  primitive.ObjectID `path:"subscriptionId" json:"subscriptionId" bson:"_id"`           // Unique Identifier of this record
	UserID          primitive.ObjectID `path:"userId"         json:"userId"         bson:"userId"`        // ID of the stream that owns this subscription
	InboxFolderID   primitive.ObjectID `path:"inboxFolderId"  json:"inboxFolderId"  bson:"inboxFolderId"` // ID of the inbox folder to put messages into
	Label           string             `path:"label"          json:"label"          bson:"label"`         // Label of this subscription
	URL             string             `path:"url"            json:"url"            bson:"url"`           // Connection URL for obtaining new sub-streams.
	Method          string             `path:"method"         json:"method"         bson:"method"`        // Method used to subscribe to remote streams (RSS, etc)
	Status          string             `path:"status"         json:"status"         bson:"status"`        // Status of the last poll of Subscription (NEW, WAITING, SUCCESS, FAILURE)
	LastPolled      int64              `path:"lastPolled"     json:"lastPolled"     bson:"lastPolled"`    // Unix Timestamp of the last date that this resource was retrieved.
	PollDuration    int                `path:"pollDuration"   json:"pollDuration"   bson:"pollDuration"`  // Time (in hours) to wait between polling this resource.
	NextPoll        int64              `path:"nextPoll"       json:"nextPoll"       bson:"nextPoll"`      // Unix Timestamp of the next time that this resource should be polled.
	journal.Journal `json:"-" bson:"journal"`
}

func NewSubscription() Subscription {
	return Subscription{
		SubscriptionID: primitive.NewObjectID(),
		PollDuration:   24, // default poll interval is 24 hours
	}
}

/*******************************************
 * DATA.OBJECT INTERFACE
 *******************************************/

// ID returns the primary key of this object
func (sub *Subscription) ID() string {
	return sub.SubscriptionID.Hex()
}

// MarkPolled updates the lastPolled and nextPoll timestamps.
func (sub *Subscription) MarkPolled() {

	// RULE: Default Poll Duration is 24 hours
	if sub.PollDuration == 0 {
		sub.PollDuration = 24
	}

	// RULE: Require that poll duration is at least 1 hour
	if sub.PollDuration < 1 {
		sub.PollDuration = 1
	}

	// Update poll time stamps
	sub.LastPolled = time.Now().Unix()
	sub.NextPoll = sub.LastPolled + int64(sub.PollDuration*60)
}
