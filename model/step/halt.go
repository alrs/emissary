package step

import (
	"github.com/benpate/rosetta/mapof"
)

// Halt represents an action-step that can update the data.DataMap custom data stored in a Stream
type Halt struct{}

// NewHalt returns a fully initialized Halt object
func NewHalt(stepInfo mapof.Any) (Halt, error) {
	return Halt{}, nil
}

// AmStep is here only to verify that this struct is a render pipeline step
func (step Halt) AmStep() {}
