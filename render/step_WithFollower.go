package render

import (
	"io"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/model/step"
	"github.com/benpate/derp"
)

// StepWithFollower represents an action-step that can update the data.DataMap custom data stored in a Stream
type StepWithFollower struct {
	SubSteps []step.Step
}

func (step StepWithFollower) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return step.execute(renderer, buffer, ActionMethodGet)
}

// Post updates the stream with approved data from the request body.
func (step StepWithFollower) Post(renderer Renderer, buffer io.Writer) PipelineBehavior {
	return step.execute(renderer, buffer, ActionMethodPost)
}

func (step StepWithFollower) execute(renderer Renderer, buffer io.Writer, actionMethod ActionMethod) PipelineBehavior {

	const location = "render.StepWithFollower.execute"

	if !renderer.IsAuthenticated() {
		return Halt().WithError(derp.NewUnauthorizedError(location, "Anonymous user is not authorized to perform this action"))
	}

	// Collect required services and values
	factory := renderer.factory()
	followerService := factory.Follower()
	followerToken := renderer.QueryParam("followerId")
	follower := model.NewFollower()
	follower.ParentID = renderer.AuthenticatedID()

	// Try to load the Follower record (unless we're creating a NEW record)
	if (followerToken != "") && (followerToken != "new") {
		if err := followerService.LoadByToken(renderer.AuthenticatedID(), followerToken, &follower); err != nil {
			if actionMethod == ActionMethodGet {
				return Halt().WithError(derp.Wrap(err, location, "Unable to load Follower via ID", followerToken))
			}
		}
	}

	// Create a new renderer tied to the Follower record
	subRenderer, err := NewModel(factory, renderer.request(), renderer.response(), followerService, &follower, renderer.template(), renderer.ActionID())

	if err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Unable to create sub-renderer"))
	}

	// Execute the render pipeline on the Follower record
	result := Pipeline(step.SubSteps).Execute(factory, subRenderer, buffer, actionMethod)
	result.Error = derp.Wrap(result.Error, location, "Error executing steps for child")

	return UseResult(result)
}
