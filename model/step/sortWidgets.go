package step

import (
	"github.com/benpate/rosetta/mapof"
)

// SortWidgets represents an action-step that can update multiple records at once
type SortWidgets struct{}

func NewSortWidgets(stepInfo mapof.Any) (SortWidgets, error) {

	return SortWidgets{}, nil
}

// AmStep is here only to verify that this struct is a render pipeline step
func (step SortWidgets) AmStep() {}
