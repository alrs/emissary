package render

import (
	"io"
)

// StepTriggerEvent represents an action-step that forwards the user to a new page.
type StepTriggerEvent struct {
	Event string
	Value string
}

func (step StepTriggerEvent) Get(renderer Renderer, _ io.Writer) PipelineBehavior {
	return nil
}

// Post updates the stream with approved data from the request body.
func (step StepTriggerEvent) Post(renderer Renderer, _ io.Writer) PipelineBehavior {
	return Continue().WithEvent(step.Event, step.Value)
}
