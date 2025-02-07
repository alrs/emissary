package render

import (
	"io"
	"strings"

	"github.com/benpate/derp"
	"github.com/benpate/form"
	"github.com/benpate/rosetta/convert"
	"github.com/benpate/rosetta/mapof"
	"github.com/benpate/table"
)

type StepTableEditor struct {
	Path string
	Form form.Element
}

func (step StepTableEditor) Get(renderer Renderer, buffer io.Writer) PipelineBehavior {

	const location = "render.StepTableEditor.Get"
	var err error

	s := renderer.schema()
	factory := renderer.factory()

	targetURL := step.getTargetURL(renderer)
	t := table.New(&s, &step.Form, renderer.object(), step.Path, factory.Icons(), targetURL)
	t.UseLookupProvider(renderer.lookupProvider())
	t.AllowAll()

	if editRow, ok := convert.IntOk(renderer.QueryParam("edit"), 0); ok {
		err = t.DrawEdit(editRow, buffer)
	} else if add := renderer.QueryParam("add"); add != "" {
		err = t.DrawAdd(buffer)
	} else {
		err = t.DrawView(buffer)
	}

	if err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error drawing table", step.Path))
	}

	return nil
}

func (step StepTableEditor) Post(renderer Renderer, _ io.Writer) PipelineBehavior {

	const location = "render.StepTableEditor.Post"

	s := renderer.schema()
	object := renderer.object()

	// Try to get the form post data
	body := mapof.NewAny()

	if err := bindBody(renderer.request(), &body); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Failed to bind body", step))
	}

	if edit := renderer.QueryParam("edit"); edit != "" {

		// Bounds checking
		editIndex, ok := convert.IntOk(edit, 0)

		if !ok {
			return Halt().WithError(derp.NewInternalError(location, "Failed to convert edit index", step.Path, edit))
		}

		if editIndex < 0 {
			return Halt().WithError(derp.NewInternalError(location, "Edit index out of range", step.Path, editIndex))
		}

		// Try to edit the row in the data table
		for _, field := range step.Form.AllElements() {
			path := step.Path + "." + edit + "." + field.Path

			if err := s.Set(object, path, body[field.Path]); err != nil {
				return Halt().WithError(derp.Wrap(err, location, "Error setting value in table", object, field.Path, path, body[field.Path]))
			}
		}

		// Try to delete an existing record
	} else if delete := renderer.QueryParam("delete"); delete != "" {

		table, err := s.Get(object, step.Path)

		if err != nil {
			return Halt().WithError(derp.Wrap(err, location, "Error locating table in data object"))
		}

		// Bounds checking
		deleteIndex, ok := convert.IntOk(delete, 0)

		if !ok {
			return Halt().WithError(derp.NewInternalError(location, "Failed to convert edit index", step.Path, edit))
		}

		if (deleteIndex < 0) || (deleteIndex >= convert.SliceLength(table)) {
			return Halt().WithError(derp.NewInternalError(location, "Edit index out of range", step.Path, deleteIndex))
		}

		// Try to find the schema element for this table control
		if ok := renderer.schema().Remove(renderer.object(), step.Path+"."+delete); !ok {
			return Halt().WithError(derp.NewInternalError(location, "Failed to remove row from table", step.Path))
		}
	}

	// Once we're done, re-render the table and send it back to the client
	targetURL := step.getTargetURL(renderer)

	factory := renderer.factory()
	t := table.New(&s, &step.Form, renderer.object(), step.Path, factory.Icons(), targetURL)
	t.UseLookupProvider(renderer.lookupProvider())
	t.AllowAll()

	if err := t.DrawView(renderer.response()); err != nil {
		return Halt().WithError(derp.Wrap(err, location, "Error building HTML"))
	}

	return nil
}

// getTargetURL returns the URL that the table should use for all of its links
func (step StepTableEditor) getTargetURL(renderer Renderer) string {
	originalPath := renderer.request().URL.Path
	actionID := renderer.ActionID()
	pathSlice := strings.Split(originalPath, "/")
	pathSlice[len(pathSlice)-1] = actionID
	return strings.Join(pathSlice, "/")
}
