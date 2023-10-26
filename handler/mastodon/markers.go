package mastodon

import (
	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/server"
	"github.com/benpate/toot/object"
	"github.com/benpate/toot/txn"
)

// https://docs.joinmastodon.org/methods/markers/
func GetMarkers(serverFactory *server.Factory) func(model.Authorization, txn.GetMarkers) (object.Marker, error) {

	return func(model.Authorization, txn.GetMarkers) (object.Marker, error) {

	}
}

func PostMarker(serverFactory *server.Factory) func(model.Authorization, txn.PostMarker) (object.Marker, error) {

	return func(model.Authorization, txn.PostMarker) (object.Marker, error) {

	}
}
