package handler

import (
	"net/http"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/render"
	"github.com/EmissarySocial/emissary/server"
	"github.com/benpate/derp"
	"github.com/labstack/echo/v4"
)

// GetStream handles GET requests
func GetStream(factoryManager *server.Factory) echo.HandlerFunc {
	return renderStream(factoryManager, render.ActionMethodGet)
}

func PostStream(factoryManager *server.Factory) echo.HandlerFunc {
	return renderStream(factoryManager, render.ActionMethodPost)
}

// renderStream is the common Stream handler for both GET and POST requests
func renderStream(factoryManager *server.Factory, actionMethod render.ActionMethod) echo.HandlerFunc {

	const location = "handler.renderStream"

	return func(ctx echo.Context) error {

		stream := model.NewStream()

		// Try to get the factory
		factory, err := factoryManager.ByContext(ctx)

		if err != nil {
			return derp.Wrap(err, location, "Unrecognized Domain")
		}

		// Try to load the stream using request data
		streamService := factory.Stream()
		streamToken := getStreamToken(ctx)

		if err := streamService.LoadByToken(streamToken, &stream); err != nil {

			// Special case: If the HOME page is missing, then this is a new database.  Forward to the admin section
			if streamToken == "home" {
				return ctx.Redirect(http.StatusTemporaryRedirect, "/startup")
			}

			return derp.Wrap(err, location, "Error loading Stream by Token", streamToken)
		}

		// Try to find the action requested by the user.  This also enforces user permissions...
		actionID := getActionID(ctx)

		if ok, err := handleJSONLD(ctx, &stream); ok {
			return derp.Wrap(err, location, "Error rendering JSON-LD")
		}

		renderer, err := render.NewStreamWithoutTemplate(factory, ctx.Request(), ctx.Response(), &stream, actionID)

		if err != nil {
			return derp.Wrap(err, location, "Error creating Renderer")
		}

		// Add webmention link header per:
		// https://www.w3.org/TR/webmention/#sender-discovers-receiver-webmention-endpoint
		if actionMethod == render.ActionMethodGet {
			ctx.Response().Header().Set("Link", "/.webmention; rel=\"webmention\"")
		}

		if err := renderHTML(factory, ctx, &renderer, actionMethod); err != nil {
			return derp.Wrap(err, location, "Error rendering page")
		}

		return nil
	}
}

// getStreamToken returns the :stream token from the Request (or a default)
func getStreamToken(ctx echo.Context) string {
	token := ctx.Param("stream")

	switch token {

	// Empty, or "zero" tokens just go to the home page instead
	case "", "000000000000000000000000":
		return "home"
	}

	return token
}
