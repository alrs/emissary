package render

import (
	"io"

	"github.com/benpate/derp"
	"github.com/benpate/exp"
	"github.com/benpate/rosetta/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StepSort represents an action-step that can update multiple records at once
type StepSort struct {
	Keys    string
	Values  string
	Message string
}

func (step StepSort) Get(renderer Renderer, _ io.Writer) PipelineBehavior {
	return nil
}

// Post updates the stream with approved data from the request body.
func (step StepSort) Post(renderer Renderer, _ io.Writer) PipelineBehavior {

	var transaction struct {
		Keys []string `form:"keys"`
	}

	// Collect form POST information
	if err := bind(renderer.request(), &transaction); err != nil {
		return Halt().WithError(derp.NewBadRequestError("render.StepSort.Post", "Error binding body"))
	}

	for index, id := range transaction.Keys {

		// Adding one so that our index does not include 0 (rank not set)
		newRank := index + 1

		// Collect inputs to make a selection criteria
		objectID, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return Halt().WithError(derp.Wrap(err, "render.StepSort.Post", "Invalid objectId", id))
		}

		criteria := exp.Equal(step.Keys, objectID)

		// Try to load the object from the database
		object, err := renderer.service().ObjectLoad(criteria)

		if err != nil {
			return Halt().WithError(derp.Wrap(err, "render.StepSort.Post", "Error loading object with criteria: ", criteria))
		}

		// If the rank for this object has not changed, then don't waste time saving it again.
		if getter, ok := object.(schema.IntGetter); ok {
			if value, ok := getter.GetIntOK(step.Values); ok {
				if value == newRank {
					continue
				}
			}
		}

		// Update the object
		if setter, ok := object.(schema.IntSetter); ok {
			if ok := setter.SetInt(step.Values, newRank); !ok {
				return Halt().WithError(derp.Wrap(err, "render.StepSort.Post", "Error updating field: ", objectID, step.Values, newRank))
			}
		}

		// Try to save back to the database
		if err := renderer.service().ObjectSave(object, step.Message); err != nil {
			return Halt().WithError(derp.Wrap(err, "render.StepSort.Post", "Error saving record tot he database", object))
		}

	}

	// Done. Do not swap on the client side.  We don't want to reload the entire page.
	return Continue().WithHeader("HX-Reswap", "none")
}
