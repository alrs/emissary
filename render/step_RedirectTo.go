package render

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/benpate/derp"
)

// StepRedirectTo represents an action-step that sends an HTTP redirect to another page.
type StepRedirectTo struct {
	URL *template.Template
}

func (step StepRedirectTo) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return step.execute(renderer)
}

// Post updates the stream with approved data from the request body.
func (step StepRedirectTo) Post(renderer Renderer, _ io.Writer) PipelineBehavior {
	return step.execute(renderer)
}

// Redirect returns an HTTP 307 Temporary Redirect that works for both GET and POST methods
func (step StepRedirectTo) execute(renderer Renderer) PipelineBehavior {

	const location = "render.StepRedirectTo.execute"
	var nextPage bytes.Buffer

	if err := step.URL.Execute(&nextPage, renderer); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error evaluating 'url'"))
	}

	if err := redirect(renderer.response(), http.StatusTemporaryRedirect, nextPage.String()); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error redirecting to new page"))
	}

	return nil
}
