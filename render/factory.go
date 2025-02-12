package render

import (
	"github.com/EmissarySocial/emissary/config"
	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/queue"
	"github.com/EmissarySocial/emissary/service"
	"github.com/EmissarySocial/emissary/tools/set"
	"github.com/benpate/form"
	"github.com/benpate/icon"
	"github.com/benpate/mediaserver"
	"github.com/benpate/steranko"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerFactory interface {
	ByDomainName(string) (Factory, error)
}

// Factory is used to locate all necessary services
type Factory interface {
	// Model Services
	ActivityStreams() *service.ActivityStreams
	Attachment() *service.Attachment
	Block() *service.Block
	Folder() *service.Folder
	Following() *service.Following
	Follower() *service.Follower
	Group() *service.Group
	Inbox() *service.Inbox
	Mention() *service.Mention
	Outbox() *service.Outbox
	Provider() *service.Provider
	Response() *service.Response
	Stream() *service.Stream
	StreamDraft() *service.StreamDraft
	Template() *service.Template
	Theme() *service.Theme
	User() *service.User
	Widget() *service.Widget

	// Other data services
	Config() config.Domain
	Content() *service.Content
	Domain() *service.Domain
	Host() string
	Hostname() string
	Icons() icon.Provider
	MediaServer() mediaserver.MediaServer
	Locator() service.Locator
	LookupProvider(primitive.ObjectID) form.LookupProvider
	OAuthClient() *service.OAuthClient
	OAuthUserToken() *service.OAuthUserToken
	Providers() set.Slice[config.Provider]
	Queue() *queue.Queue
	Steranko() *steranko.Steranko
	StreamUpdateChannel() chan model.Stream
}
