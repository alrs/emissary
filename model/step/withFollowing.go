package step

import (
	"github.com/benpate/derp"
	"github.com/benpate/rosetta/convert"
	"github.com/benpate/rosetta/mapof"
)

// WithFollowing represents an action-step that can update the data.DataMap custom data stored in a Stream
type WithFollowing struct {
	SubSteps []Step
}

// NewWithFollowing returns a fully initialized WithFollowing object
func NewWithFollowing(stepInfo mapof.Any) (WithFollowing, error) {

	const location = "NewWithFollowing"

	subSteps, err := NewPipeline(convert.SliceOfMap(stepInfo["steps"]))

	if err != nil {
		return WithFollowing{}, derp.Wrap(err, location, "Invalid 'steps'", stepInfo)
	}

	return WithFollowing{
		SubSteps: subSteps,
	}, nil
}

// AmStep is here only to verify that this struct is a render pipeline step
func (step WithFollowing) AmStep() {}
