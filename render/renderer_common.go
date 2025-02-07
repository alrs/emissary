package render

import (
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/benpate/domain"
	"github.com/benpate/exp"
	"github.com/benpate/form"
	"github.com/benpate/hannibal/streams"
	"github.com/benpate/html"
	"github.com/benpate/rosetta/convert"
	"github.com/benpate/rosetta/list"
	"github.com/benpate/rosetta/mapof"
	"github.com/benpate/rosetta/sliceof"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Common provides common rendering functions that are needed by ALL renderers
type Common struct {
	_factory       Factory             // Factory interface is required for locating other services.
	_request       *http.Request       // Pointer to the HTTP request we are serving
	_response      http.ResponseWriter // ResponseWriter for this request
	_authorization model.Authorization // Authorization information for the current user
	_template      model.Template      // Template to use for this renderer
	action         model.Action        // Action to be performed on the Template
	actionID       string              // Token that identifies the action requested in the URL

	requestData mapof.Any // Temporary data scope for this request

	// Cached values, do not populate unless needed
	domain model.Domain // This is a value because we expect to use it in every request.
}

func NewCommon(factory Factory, request *http.Request, response http.ResponseWriter, template model.Template, actionID string) (Common, error) {

	const location = "render.NewCommon"

	// Retrieve the user's authorization information
	steranko := factory.Steranko()
	authorization := getAuthorization(steranko, request)

	// Verify that the actionID is valid
	action, ok := template.Actions[actionID]

	if !ok {
		return Common{}, derp.NewBadRequestError(location, "Invalid action", actionID)
	}

	// Return a new Common renderer
	return Common{
		_factory:       factory,
		_request:       request,
		_response:      response,
		_authorization: authorization,
		_template:      template,
		action:         action,
		actionID:       actionID,
		requestData:    mapof.NewAny(),
		domain:         model.NewDomain(),
	}, nil
}

/******************************************
 * Renderer Interface
 ******************************************/

// context returns request context embedded in this renderer.
func (w Common) factory() Factory {
	return w._factory
}

func (w Common) request() *http.Request {
	return w._request
}

func (w Common) response() http.ResponseWriter {
	return w._response
}

func (w Common) authorization() model.Authorization {
	return w._authorization
}

// Action returns the model.Action configured into this renderer
func (w Common) Action() model.Action {
	return w.action
}

func (w Common) ActionID() string {
	return w.actionID
}

/******************************************
 * Page Defaults
 ******************************************/

func (w Common) PageTitle() string {
	return ""
}

func (w Common) Summary() string {
	return ""
}

/******************************************
 * Request Info
 ******************************************/

// Returns the request method
func (w Common) Method() string {
	return w._request.Method
}

// Host returns the protocol + the Hostname
func (w Common) Host() string {
	return w.Protocol() + w.Hostname()
}

// URL returns the originally requested URL
func (w Common) URL() string {
	return w._request.URL.RequestURI()
}

// Protocol returns http:// or https:// used for this request
func (w Common) Protocol() string {
	return domain.Protocol(w.Hostname())
}

// Hostname returns the configured hostname for this request
func (w Common) Hostname() string {
	return w._request.Host
}

// Path returns the request path
func (w Common) Path() string {
	return w._request.URL.Path
}

func (w Common) PathList() list.List {
	return list.BySlash(w.Path()).Tail()
}

func (w Common) SetQueryParam(name string, value string) {
	query := w._request.URL.Query()
	query.Set(name, value)
	w._request.URL.RawQuery = query.Encode()
}

// Returns the designated request parameter
func (w Common) QueryParam(param string) string {
	urlValue := w._request.URL.Query()[param]
	// NOTE: we're getting the QueryParam value this way because the context.QueryParam() method
	// doesn't seem to get updates when we change the URL via the SetQueryParam() step.
	return convert.String(urlValue)
}

// IsPartialRequest returns TRUE if this is a partial page request from htmx.
func (w Common) IsPartialRequest() bool {
	return w._request.Header.Get("HX-Request") != ""
}

// templateRole returns the the role that the current Template performs in the system.
// Used for selecting eligible child templates.
func (w Common) templateRole() string {
	return ""
}

// UserCan returns TRUE if the current user has the specified permission.
// Default implementation returns FALSE for all requests.
func (w Common) UserCan(_ string) bool {
	return false
}

// Now returns the current time in milliseconds since the Unix epoch
func (w Common) Now() int64 {
	return time.Now().Unix()
}

// NavigationID returns the the identifier of the top-most stream in the
// navigation.  The "common" renderer just returns a default value that
// other renderers should override.
func (w Common) NavigationID() string {
	return ""
}

/***************************
 * Request Data (Getters and Setters)
 **************************/

func (w Common) GetBool(name string) bool {
	return w.requestData.GetBool(name)
}

func (w Common) GetContent() template.HTML {
	return w.GetHTML("content")
}

func (w Common) GetFloat(name string) float64 {
	return w.requestData.GetFloat(name)
}

func (w Common) GetHTML(name string) template.HTML {
	return template.HTML(w.requestData.GetString(name))
}

func (w Common) GetInt(name string) int {
	return w.requestData.GetInt(name)
}

func (w Common) GetInt64(name string) int64 {
	return w.requestData.GetInt64(name)
}

func (w Common) GetString(name string) string {
	return w.requestData.GetString(name)
}

func (w Common) SetBool(name string, value bool) {
	w.requestData.SetBool(name, value)
}

func (w Common) SetContent(value string) {
	w.SetString("content", value)
}

func (w Common) SetFloat(name string, value float64) {
	w.requestData.SetFloat(name, value)
}

func (w Common) SetInt(name string, value int) {
	w.requestData.SetInt(name, value)
}

func (w Common) SetInt64(name string, value int64) {
	w.requestData.SetInt64(name, value)
}

func (w Common) SetString(name string, value string) {
	w.requestData.SetString(name, value)
}

/******************************************
 * Domain Data
 ******************************************/

func (w Common) DomainLabel() (string, error) {
	if domain, err := w.getDomain(); err != nil {
		return "", err
	} else {
		return domain.Label, nil
	}
}

func (w Common) DomainHasSignupForm() (bool, error) {
	if domain, err := w.getDomain(); err != nil {
		return false, err
	} else {
		return domain.HasSignupForm(), nil
	}
}

/***************************
 * Access Permissions
 **************************/

// IsAuthenticated returns TRUE if the user is signed in
func (w Common) IsAuthenticated() bool {
	authorization := w.authorization()
	return authorization.IsAuthenticated()
}

// IsOwner returns TRUE if the user is a Domain Owner
func (w Common) IsOwner() bool {
	authorization := w.authorization()
	return authorization.DomainOwner
}

// AuthenticatedID returns the unique ID of the currently logged in user (may be nil).
func (w Common) AuthenticatedID() primitive.ObjectID {
	authorization := w.authorization()
	return authorization.UserID
}

// UserName returns the DisplayName of the user
func (w Common) UserName() (string, error) {
	user, err := w.getUser()

	if err != nil {
		return "", derp.Wrap(err, "render.Stream.UserName", "Error loading User")
	}

	return user.DisplayName, nil
}

func (w Common) Avatar(url string, size int) template.HTML {
	b := html.New()
	b.Empty("img").Attr("src", url).Style("width:"+convert.String(size)+"px", "border-radius:"+convert.String(size)+"px").Close()
	return template.HTML(b.String())
}

// UserAvatar returns the avatar image of the user
func (w Common) UserImage() (string, error) {
	user, err := w.getUser()

	if err != nil {
		return "", derp.Wrap(err, "render.Stream.UserAvatar", "Error loading User")
	}

	return user.ActivityPubAvatarURL(), nil
}

/******************************************
 * Misc Helper Methods
 ******************************************/

func (w Common) ActivityStream(uri string) streams.Document {
	result, err := w._factory.ActivityStreams().Load(uri)
	derp.Report(err)
	return result
}

func (w Common) lookupProvider() form.LookupProvider {

	userID := w.AuthenticatedID()
	return w._factory.LookupProvider(userID)
}

func (w Common) executeTemplate(writer io.Writer, name string, data any) error {
	if err := w._template.HTMLTemplate.ExecuteTemplate(writer, name, data); err != nil {
		return derp.Wrap(err, "render.Common.executeTemplate", "Error executing template", name)
	}

	return nil
}

// withViewPermission augments a query criteria to include the
// group authorizations of the currently signed in user.
func (w Common) withViewPermission(criteria exp.Expression) exp.Expression {

	result := criteria.
		And(exp.Equal("deleteDate", 0)) // Stream must not be deleted

	// If the user IS NOT a domain owner, then we must also
	// check their permission to VIEW this stream
	authorization := w.authorization()

	if !authorization.DomainOwner {
		result = result.And(exp.In("defaultAllow", authorization.AllGroupIDs())).
			And(exp.LessThan("publishDate", time.Now().Unix())) // Stream must be published
	}

	return result
}

func (w Common) template() model.Template {
	return w._template
}

// getUser loads/caches the currently-signed-in user to be used by other functions in this renderer
func (w Common) getUser() (model.User, error) {

	userService := w._factory.User()
	steranko := w._factory.Steranko()
	result := model.NewUser()
	authorization := getAuthorization(steranko, w._request)

	if err := userService.LoadByID(authorization.UserID, &result); err != nil {
		return model.User{}, derp.Wrap(err, "render.Common.getUser", "Error loading user from database", authorization.UserID)
	}

	return result, nil
}

// getDomain retrieves the current domain model object from the domain service cache
func (w *Common) getDomain() (model.Domain, error) {

	domainService := w._factory.Domain()

	if !domainService.Ready() {
		return model.Domain{}, derp.NewInternalError("render.Common.getDomain", "Domain service not ready", nil)
	}

	return domainService.Get(), nil
}

/******************************************
 * Global Queries
 ******************************************/

// Navigation returns an array of Streams that have a Zero ParentID
func (w Common) Navigation() (sliceof.Object[model.StreamSummary], error) {
	criteria := w.withViewPermission(exp.Equal("parentId", primitive.NilObjectID))
	builder := NewQueryBuilder[model.StreamSummary](w._factory.Stream(), criteria)

	result, err := builder.Top60().ByRank().Slice()
	return result, err
}

/******************************************
 * Additional Data
 ******************************************/

// AdminSections returns labels and values for all hard-coded sections of the administrator area.
func (w Common) AdminSections() []form.LookupCode {
	return []form.LookupCode{
		{
			Value: "domain",
			Label: "General",
		},
		{
			Value: "navigation",
			Label: "Navigation",
		},
		/* REMOVING EXTERNAL CONNECTIONS UNTIL THEY'RE NEEDED
		{
			Value: "connections",
			Label: "Services",
		},
		*/
		{
			Value: "groups",
			Label: "Groups",
		},
		{
			Value: "users",
			Label: "Users",
		},
		{
			Value: "blocks",
			Label: "Blocks",
		},
	}
}
