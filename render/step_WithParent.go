package render

import (
	"io"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/model/step"
	"github.com/benpate/derp"
)

// StepWithParent represents an action-step that can update the data.DataMap custom data stored in a Stream
type StepWithParent struct {
	SubSteps []step.Step
}

func (step StepWithParent) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return nil
}

// Post executes the subSteps on the parent Stream
func (step StepWithParent) Post(renderer Renderer, buffer io.Writer) PipelineBehavior {

	const location = "render.StepWithParent.Post"

	var parent model.Stream

	factory := renderer.factory()
	streamRenderer := renderer.(*Stream)

	if err := factory.Stream().LoadByID(streamRenderer._stream.ParentID, &parent); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error listing parent"))
	}

	// Make a renderer with the new parent stream
	// TODO: LOW: Is "view" really the best action to use here??
	parentStream, err := NewStreamWithoutTemplate(streamRenderer.factory(), streamRenderer.request(), streamRenderer.response(), &parent, "")

	if err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error creating renderer for parent"))
	}

	// Execute the POST render pipeline on the parent
	result := Pipeline(step.SubSteps).Post(factory, &parentStream, buffer)
	result.Error = derp.Wrap(result.Error, location, "Error executing steps for parent")
	return UseResult(result)
}
