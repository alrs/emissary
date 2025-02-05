package render

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/service"
	"github.com/benpate/data"
	"github.com/benpate/derp"
	"github.com/benpate/exp"
	builder "github.com/benpate/exp-builder"
	"github.com/benpate/rosetta/schema"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a renderer for the admin/users page
// It can only be accessed by a Domain Owner
type User struct {
	_user *model.User
	Common
}

// NewUser returns a fully initialized `User` renderer.
func NewUser(factory Factory, request *http.Request, response http.ResponseWriter, template model.Template, user *model.User, actionID string) (User, error) {

	const location = "render.NewGroup"

	// Create the underlying Common renderer
	common, err := NewCommon(factory, request, response, template, actionID)

	if err != nil {
		return User{}, derp.Wrap(err, location, "Error creating common renderer")
	}

	// Verify that the user is a Domain Owner
	if !common._authorization.DomainOwner {
		return User{}, derp.NewForbiddenError(location, "Must be domain owner to continue")
	}

	// Return the User renderer
	return User{
		_user:  user,
		Common: common,
	}, nil
}

/******************************************
 * RENDERER INTERFACE
 ******************************************/

// Render generates the string value for this User
func (w User) Render() (template.HTML, error) {

	var buffer bytes.Buffer

	// Execute step (write HTML to buffer, update context)
	status := Pipeline(w.action.Steps).Get(w.factory(), &w, &buffer)

	if status.Error != nil {
		err := derp.Wrap(status.Error, "render.User.Render", "Error generating HTML")
		derp.Report(err)
		return "", err
	}

	// Success!
	status.Apply(w._response)
	return template.HTML(buffer.String()), nil
}

// View executes a separate view for this User
func (w User) View(actionID string) (template.HTML, error) {

	renderer, err := NewUser(w._factory, w._request, w._response, w._template, w._user, actionID)

	if err != nil {
		return template.HTML(""), derp.Wrap(err, "render.User.View", "Error creating renderer")
	}

	return renderer.Render()
}

func (w User) NavigationID() string {
	return "admin"
}

func (w User) Token() string {
	return "users"
}

func (w User) PageTitle() string {
	return "Settings"
}

func (w User) Permalink() string {
	return ""
}

func (w User) object() data.Object {
	return w._user
}

func (w User) objectID() primitive.ObjectID {
	return w._user.UserID
}

func (w User) objectType() string {
	return "User"
}

func (w User) schema() schema.Schema {
	return schema.New(model.UserSchema())
}

func (w User) service() service.ModelService {
	return w._factory.User()
}

func (w User) executeTemplate(writer io.Writer, name string, data any) error {
	return w._template.HTMLTemplate.ExecuteTemplate(writer, name, data)
}

func (w User) clone(action string) (Renderer, error) {
	return NewUser(w._factory, w._request, w._response, w._template, w._user, action)
}

/******************************************
 * Domain Data
 ******************************************/

func (w User) SignupForm() model.SignupForm {
	return w._factory.Domain().Get().SignupForm
}

/******************************************
 * User Data
 ******************************************/

func (w User) UserID() string {
	return w._user.UserID.Hex()
}

func (w User) Label() string {
	return w._user.DisplayName
}

func (w User) DisplayName() string {
	return w._user.DisplayName
}

func (w User) ImageURL() string {
	return w._user.ActivityPubAvatarURL()
}

/******************************************
 * Query Builders
 ******************************************/

func (w User) Users() *QueryBuilder[model.UserSummary] {

	query := builder.NewBuilder().
		String("displayName").
		ObjectID("groupId")

	criteria := exp.And(
		query.Evaluate(w._request.URL.Query()),
		exp.Equal("deleteDate", 0),
	)

	result := NewQueryBuilder[model.UserSummary](w._factory.User(), criteria)

	return &result
}

/******************************************
 * ADDITIONAL DATA
 ******************************************/

// AssignedGroups lists all groups to which the current user is assigned.
func (w User) AssignedGroups() ([]model.Group, error) {
	groupService := w.factory().Group()
	result, err := groupService.ListByIDs(w._user.GroupIDs...)

	return result, derp.Wrap(err, "render.User.AssignedGroups", "Error listing groups", w._user.GroupIDs)
}

func (w User) debug() {
	log.Debug().Interface("object", w.object()).Msg("renderer_admin_users")
}
