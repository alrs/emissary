package step

import (
	"github.com/benpate/derp"
	"github.com/benpate/form"
	"github.com/benpate/rosetta/maps"
)

// EditModelObject is an action that can add new sub-streams to the domain.
type EditModelObject struct {
	Form     form.Element
	Defaults []Step
}

// NewEditModelObject returns a fully initialized EditModelObject record
func NewEditModelObject(stepInfo maps.Map) (EditModelObject, error) {

	// Parse form
	f, err := form.Parse(stepInfo.GetInterface("form"))

	if err != nil {
		return EditModelObject{}, derp.Wrap(err, "model.step.NewEditModelObject", "Invalid 'form'", stepInfo)
	}

	// Parse defaults
	defaults, err := NewPipeline(stepInfo.GetSliceOfMap("defaults"))

	if err != nil {
		return EditModelObject{}, derp.Wrap(err, "model.step.NewEditModelObject", "Invalid 'defaults'", stepInfo)
	}

	return EditModelObject{
		Form:     f,
		Defaults: defaults,
	}, nil
}

// AmStep is here only to verify that this struct is a render pipeline step
func (step EditModelObject) AmStep() {}
