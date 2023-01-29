package service

import (
	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/benpate/digit"
	"github.com/benpate/remote"
	"github.com/benpate/rosetta/first"
	"github.com/labstack/gommon/random"
)

func (service *Following) connect_WebSub(following *model.Following, hub digit.Link) error {

	const location = "service.Following.ConnectWebSub"

	var success string
	var failure string

	// Autocompute the topic.  Use "self" link first, or just the following URL
	self := following.GetLink("rel", model.LinkRelationSelf)

	// Update values in the following object
	following.Method = model.FollowMethodWebSub
	following.URL = first.String(self.Href, following.URL)
	following.Secret = random.String(32)
	following.PollDuration = 30

	// "Pending" status means that we're still waiting on the WebSub connection
	if err := service.SetStatus(following, model.FollowingStatusPending, ""); err != nil {
		return derp.Wrap(err, location, "Error updating following status", following)
	}

	// Send request to the hub
	transaction := remote.Post(hub.Href).
		Header("Accept", followingMimeStack).
		Form("hub.mode", "subscribe").
		Form("hub.topic", following.URL).
		Form("hub.callback", service.websubCallbackURL(following)).
		Form("hub.secret", following.Secret).
		Form("hub.lease_seconds", "2582000").
		Response(&success, &failure)

	if err := transaction.Send(); err != nil {
		return derp.Wrap(err, location, "Error sending WebSub subscription request", hub.Href)
	}

	// Success!
	return nil
}

func (service *Following) disconnect_WebSub(following *model.Following) {

	// Find the "hub" link for this following
	for _, link := range following.Links {
		if link.RelationType == "hub" {

			transaction := remote.Post(link.Href).
				Form("hub.mode", "unsubscribe").
				Form("hub.topic", following.URL).
				Form("hub.callback", service.websubCallbackURL(following))

			transaction.Send() // Silent fail is okay here.
		}
	}
}

func (service *Following) websubCallbackURL(following *model.Following) string {
	return service.host + "/.websub/" + following.UserID.Hex() + "/" + following.FollowingID.Hex()
}
