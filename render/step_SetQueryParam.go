package render

import (
	"io"
	"text/template"
)

// StepSetQueryParam represents an action-step that sets values to the request query string
type StepSetQueryParam struct {
	Values map[string]*template.Template
}

// Get displays a form where users can update stream data
func (step StepSetQueryParam) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return step.Do(renderer)
}

// Post updates the stream with approved data from the request body.
func (step StepSetQueryParam) Post(renderer Renderer, _ io.Writer) PipelineBehavior {
	return step.Do(renderer)
}

func (step StepSetQueryParam) Do(renderer Renderer) PipelineBehavior {
	query := renderer.request().URL.Query()

	for key, value := range step.Values {
		queryValue := executeTemplate(value, renderer)
		query.Set(key, queryValue)
	}

	renderer.request().URL.RawQuery = query.Encode()
	return nil
}
