package render

import (
	"io"

	"github.com/benpate/derp"
	"github.com/benpate/rosetta/schema"
)

// StepSetState represents an action-step that can change a Stream's state
type StepSetState struct {
	StateID string
}

func (step StepSetState) Get(renderer Renderer, _ io.Writer) error {
	return nil
}

func (step StepSetState) UseGlobalWrapper() bool {
	return true
}

// Post updates the stream with configured data, and moves the stream to a new state
func (step StepSetState) Post(renderer Renderer, _ io.Writer) error {

	// Try to set the state via the Path interface.
	object := renderer.object()
	if setter, ok := object.(schema.StringSetter); ok {
		if ok := setter.SetString("stateId", step.StateID); !ok {
			return derp.NewInternalError("Unable to set stateId", step.StateID)
		}
	}

	return nil
}
