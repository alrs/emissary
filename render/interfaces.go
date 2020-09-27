package render

import (
	"github.com/benpate/data"
	"github.com/benpate/ghost/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplateService interface {
	Load(string) (*model.Template, error)
}

type StreamService interface {
	LoadByToken(string) (*model.Stream, error)
	LoadParent(*model.Stream) (*model.Stream, error)
	ListByParent(primitive.ObjectID) (data.Iterator, error)
}

// Renderer wraps the Render method, which returns an HTML rendering of an object.
type Renderer interface {
	// Render returns an HTML rendering of an object.
	Render() (string, error)

	Stream() StreamWrapper
}
