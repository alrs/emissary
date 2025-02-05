package render

import (
	"io"

	"github.com/benpate/derp"
	"github.com/benpate/form"
	"github.com/benpate/rosetta/mapof"
)

// StepEditProperties represents an action-step that can edit/update Container in a streamDraft.
type StepEditProperties struct {
	Title string
	Paths []string
}

func (step StepEditProperties) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {

	schema := renderer.schema()
	streamRenderer := renderer.(*Stream)
	stream := streamRenderer._stream

	element := form.Element{
		Type:     "layout-vertical",
		Label:    step.Title,
		Children: []form.Element{},
	}

	for _, path := range step.Paths {

		switch path {

		case "token":
			element.Children = append(element.Children,
				form.Element{
					Path:        path,
					Type:        "text",
					Label:       "URL Token",
					Options:     mapof.Any{"format": "token"},
					Description: "Human-friendly web address",
				})

		case "label":
			element.Children = append(element.Children,
				form.Element{
					Path:        path,
					Type:        "text",
					Label:       "Label",
					Description: "Displayed on navigation, pages, and indexes",
					Options:     mapof.Any{"maxlength": 100},
				})

		case "description":

			element.Children = append(element.Children,
				form.Element{
					Type:        "textarea",
					Path:        path,
					Label:       "Text Description",
					Description: "Long description displays on pages and indexes",
					Options:     mapof.Any{"maxlength": 1000},
				})

		}
	}

	// Create HTML for the form
	html, err := form.Editor(schema, element, stream, renderer.lookupProvider())

	if err != nil {
		return Halt().WithError(derp.Wrap(err, "render.StepEditProperties.Get", "Error generating form HTML"))
	}

	result := WrapModalForm(renderer.response(), renderer.URL(), html)
	if _, err = buffer.Write([]byte(result)); err != nil {
		return Halt().WithError(derp.Wrap(err, "render.StepEditProperties.Get", "Error writing response"))
	}

	return nil
}

func (step StepEditProperties) Post(renderer Renderer, _ io.Writer) PipelineBehavior {

	const location = "render.StepEditProperties.Post"
	body := mapof.NewAny()

	if err := bind(renderer.request(), &body); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error binding request body"))
	}

	schema := renderer.schema()
	streamRenderer := renderer.(*Stream)
	stream := streamRenderer._stream

	for _, path := range step.Paths {
		if value, ok := body[path]; ok {
			if err := schema.Set(stream, path, value); err != nil {
				return Halt().WithError(derp.Wrap(err, location, "Error setting value", path, body[path]))
			}
		}
	}

	if err := schema.Validate(stream); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error validating data", stream))
	}

	// Success!
	return Continue().WithEvent("closeModal", "true")
}
