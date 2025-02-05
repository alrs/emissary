package render

import (
	"encoding/json"
	"io"

	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	accept "github.com/timewasted/go-accept-headers"
)

// StepViewJSONLD represents an action-step that can render a Stream into HTML
type StepViewJSONLD struct {
	Method string
}

// Get renders the Stream HTML to the context
func (step StepViewJSONLD) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {

	if step.Method != "post" {
		return step.execute(renderer, buffer)
	}

	return nil
}

func (step StepViewJSONLD) Post(renderer Renderer, buffer io.Writer) PipelineBehavior {

	if step.Method != "get" {
		return step.execute(renderer, buffer)
	}

	return nil
}

func (step StepViewJSONLD) execute(renderer Renderer, buffer io.Writer) PipelineBehavior {

	// Try to negotiate the correct content type
	acceptHeader := renderer.request().Header.Get("Accept")
	accept, err := accept.Negotiate(acceptHeader, model.MimeTypeHTML, model.MimeTypeActivityPub, model.MimeTypeJSONLD, model.MimeTypeJSON)

	// If there is an error in content negotiation, then no JSON-LD for you
	if err != nil {
		return nil
	}

	// If you haven't requested ActivityPub, JSON-LD, or JSON, then no JSON-LD for you
	switch accept {
	case model.MimeTypeActivityPub:
	case model.MimeTypeJSONLD:
	case model.MimeTypeJSON:
	default:
		return nil
	}

	// JSON-LD FOR YOU!!!!

	// Now, try to get a JSONLDGetter from the renderer
	if getter, ok := renderer.object().(model.JSONLDGetter); ok {

		// Write the object as JSON
		result, err := json.Marshal(getter.GetJSONLD())

		if err != nil {
			return Halt().WithError(derp.Wrap(err, "render.StepViewJSONLD.Get", "Error marshalling JSONLD"))
		}

		// Write the JSON to the output buffer
		if _, err := buffer.Write(result); err != nil {
			return Halt().WithError(derp.Wrap(err, "render.StepViewJSONLD.Get", "Error writing JSONLD to buffer"))
		}

		// Done.  Return result as pure JSON.
		return Halt().AsFullPage().WithContentType(model.MimeTypeActivityPub)
	}

	// If you're here, that means the template designer used step on a non-JSONLDGetter object type.  Shame.
	return Halt().WithError(derp.NewNotFoundError("render.StepViewJSONLD.Get", "Object does not implement JSONLDGetter interface"))
}
