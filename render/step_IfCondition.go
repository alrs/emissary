package render

import (
	"bytes"
	"html/template"
	"io"
	"strings"

	"github.com/EmissarySocial/emissary/model/step"
	"github.com/benpate/derp"
)

// StepIfCondition represents an action-step that can update the data.DataMap custom data stored in a Stream
type StepIfCondition struct {
	Condition *template.Template
	Then      []step.Step
	Otherwise []step.Step
}

// Get displays a form where users can update stream data
func (step StepIfCondition) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return step.execute(renderer, buffer, ActionMethodGet)
}

// Post updates the stream with approved data from the request body.
func (step StepIfCondition) Post(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return step.execute(renderer, buffer, ActionMethodPost)
}

// Get displays a form where users can update stream data
func (step StepIfCondition) execute(renderer Renderer, buffer io.Writer, method ActionMethod) PipelineBehavior {

	const location = "renderer.StepIfCondition.execute"

	factory := renderer.factory()

	if step.evaluateCondition(renderer) {
		result := Pipeline(step.Then).Execute(factory, renderer, buffer, method)
		result.Error = derp.Wrap(result.Error, location, "Error executing 'then' sub-steps")
		return UseResult(result)
	}

	result := Pipeline(step.Otherwise).Get(factory, renderer, buffer)
	result.Error = derp.Wrap(result.Error, location, "Error executing 'otherwise' sub-steps")
	return UseResult(result)
}

// evaluateCondition executes the conditional template and
func (step StepIfCondition) evaluateCondition(renderer Renderer) bool {

	var result bytes.Buffer

	if err := step.Condition.Execute(&result, renderer); err != nil {
		return false
	}

	return (strings.TrimSpace(result.String()) == "true")
}
