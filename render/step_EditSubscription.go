package render

import (
	"io"
	"net/http"

	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/benpate/form"
	"github.com/benpate/rosetta/maps"
	"github.com/benpate/rosetta/null"
	"github.com/benpate/rosetta/schema"
)

// StepEditSubscription is an action that can edit a subscription for the current user.
type StepEditSubscription struct {
}

func (step StepEditSubscription) Get(renderer Renderer, buffer io.Writer) error {

	// Requre that users are signed in to use this modal
	if !renderer.IsAuthenticated() {
		return derp.NewUnauthorizedError("render.StepAddSubscription", "User is not authenticated", nil)
	}

	// Get the request context
	context := renderer.context()

	// Load the existing subscription
	subscriptionService := renderer.factory().Subscription()
	subscription := model.NewSubscription()

	if err := subscriptionService.LoadByToken(renderer.AuthenticatedID(), context.QueryParam("subscriptionId"), &subscription); err != nil {
		return derp.Wrap(err, "render.StepEditSubscription", "Error loading subscription")
	}

	// Create a new form
	f := step.form()
	html, err := f.Editor(&subscription, nil)

	if err != nil {
		return derp.Wrap(err, "render.StepAddSubscription", "Error creating form editor", nil)
	}

	// Wrap the form as a modal dialog (with submit buttons)
	html = WrapModalForm(context.Response(), renderer.URL(), html)

	// Done.
	return context.HTML(http.StatusOK, html)

}

func (step StepEditSubscription) UseGlobalWrapper() bool {
	return false
}

func (step StepEditSubscription) Post(renderer Renderer) error {

	var transaction struct {
		URL          string `form:"url"          path:"url"`
		PollDuration int    `form:"pollDuration" path:"pollDuration"`
	}

	// Requre that users are signed in to use this modal
	if !renderer.IsAuthenticated() {
		return derp.NewUnauthorizedError("render.StepAddSubscription", "User is not authenticated", nil)
	}

	// Collect data from the form POST
	context := renderer.context()

	if err := context.Bind(&transaction); err != nil {
		return derp.Wrap(err, "render.StepAddSubscription", "Error reading form data", nil)
	}

	if err := step.form().Schema.Validate(transaction); err != nil {
		return derp.Wrap(err, "render.StepAddSubscription", "Subscription Data is invalid", transaction)
	}

	// Load the existing subscription
	subscriptionService := renderer.factory().Subscription()
	subscription := model.NewSubscription()

	if err := subscriptionService.LoadByToken(renderer.AuthenticatedID(), context.QueryParam("subscriptionId"), &subscription); err != nil {
		return derp.Wrap(err, "render.StepEditSubscription", "Error loading subscription")
	}

	subscription.URL = transaction.URL
	subscription.PollDuration = transaction.PollDuration

	// Save the subscription to the database
	if err := subscriptionService.Save(&subscription, "Updated by User"); err != nil {
		return derp.Wrap(err, "render.StepAddSubscription", "Error saving subscription", subscription)
	}

	// Close the Modal Dialog and return
	CloseModal(context, "")
	return context.NoContent(http.StatusOK)
}

func (step StepEditSubscription) form() form.Form {
	return form.Form{
		Schema: schema.New(schema.Object{
			Properties: schema.ElementMap{
				"url":          schema.String{Format: "uri", MaxLength: 512, Required: true},
				"pollDuration": schema.Integer{Default: null.NewInt64(24), Minimum: null.NewInt64(1), Maximum: null.NewInt64(24 * 30), Required: true},
			},
		}),
		Element: form.Element{
			Type:  "layout-vertical",
			Label: "Edit Subscription Details",
			Children: []form.Element{
				{
					Type:        "text",
					Label:       "Website URL",
					Path:        "url",
					Description: "Enter the URL of the website you want to subscribe to.",
				},
				{
					Type:  "select",
					Label: "Frequency",
					Path:  "pollDuration",
					Options: maps.Map{
						"enum": []form.LookupCode{
							{Value: "1", Label: "Hourly"},
							{Value: "6", Label: "Every 6 Hours"},
							{Value: "12", Label: "Every 12 Hours"},
							{Value: "24", Label: "Once per Day"},
							{Value: "168", Label: "Once per Week"},
							{Value: "720", Label: "Once per Month"},
						},
					},
				},
			},
		},
	}
}
