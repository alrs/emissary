package handler

import (
	"net/http"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/server"
	"github.com/benpate/derp"
	"github.com/benpate/rosetta/mapof"
	"github.com/labstack/echo/v4"
)

// GetSignIn generates an echo.HandlerFunc that handles GET /signin requests
func GetSignIn(serverFactory *server.Factory) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		// Locate the current domain
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.NewInternalError("handler.PostSignIn", "Invalid Domain.")
		}

		// Get the standard Signin page
		template := factory.Domain().Theme().HTMLTemplate

		// Get a clean version of the URL query parameters
		queryString := cleanQueryParams(ctx.QueryParams())

		// Render the template
		if err := template.ExecuteTemplate(ctx.Response(), "signin", queryString); err != nil {
			return derp.Wrap(err, "handler.GetSignIn", "Error executing template")
		}

		return nil
	}
}

// PostSignIn generates an echo.HandlerFunc that handles POST /signin requests
func PostSignIn(serverFactory *server.Factory) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		// Locate the current domain
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.NewInternalError("handler.PostSignIn", "Invalid Domain.")
		}

		s := factory.Steranko()

		if err := s.SignIn(ctx); err != nil {
			ctx.Response().Header().Add("HX-Trigger", "SigninError")
			return ctx.HTML(derp.ErrorCode(err), derp.Message(err))
		}

		ctx.Response().Header().Add("HX-Trigger", "SigninSuccess")

		return ctx.NoContent(200)
	}
}

// PostSignOut generates an echo.HandlerFunc that handles POST /signout requests
func PostSignOut(serverFactory *server.Factory) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.NewInternalError("handler.PostSignOut", "Invalid Request.  Please try again later.")
		}

		s := factory.Steranko()

		if err := s.SignOut(ctx); err != nil {
			return derp.Wrap(err, "handler.PostSignOut", "Error Signing Out")
		}

		// Forward the user back to the home page of the website.
		ctx.Response().Header().Add("HX-Redirect", "/")
		return ctx.NoContent(http.StatusNoContent)
	}
}

func GetResetPassword(serverFactory *server.Factory) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return executeDomainTemplate(serverFactory, ctx, "reset-password")
	}
}

func PostResetPassword(serverFactory *server.Factory) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		var transaction struct {
			EmailAddress string `form:"emailAddress"`
		}

		// Try to get the POST transaction data from the request body
		if err := ctx.Bind(&transaction); err != nil {
			return derp.Wrap(err, "handler.PostResetPassword", "Error binding form data")
		}

		// Try to get the factory for this domain
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.NewInternalError("handler.GetResetCode", "Invalid domain")
		}

		// Try to load the user by username.  If the user cannot be found, the response
		// will still be sent.
		userService := factory.User()
		user := model.NewUser()

		if err := userService.LoadByUsernameOrEmail(transaction.EmailAddress, &user); err == nil {
			userService.SendPasswordResetEmail(&user)
		}

		// Return a success message regardless of whether or not the user was found.
		template := factory.Domain().Theme().HTMLTemplate

		if err := template.ExecuteTemplate(ctx.Response(), "reset-confirm", nil); err != nil {
			return derp.Wrap(err, "handler.GetResetCode", "Error executing template")
		}

		return nil
	}
}

// GetResetCode displays a form (authenticated by the reset code) for resetting a user's password
func GetResetCode(serverFactory *server.Factory) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		// Try to get the factory for this domain
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.NewInternalError("handler.GetResetCode", "Invalid domain")
		}

		// Try to load the user by userID and resetCode
		userService := factory.User()

		user := model.NewUser()
		userID := ctx.QueryParam("userId")
		resetCode := ctx.QueryParam("code")

		if err := userService.LoadByResetCode(userID, resetCode, &user); err != nil {
			return derp.Wrap(err, "handler.GetResetCode", "Error loading user")
		}

		// Try to render the HTML response
		template := factory.Domain().Theme().HTMLTemplate

		object := mapof.Any{
			"userId":      userID,
			"displayName": user.DisplayName,
			"code":        resetCode,
		}

		if err := template.ExecuteTemplate(ctx.Response(), "reset-code", object); err != nil {
			return derp.Wrap(err, "handler.GetResetCode", "Error executing template")
		}

		return nil
	}
}

func PostResetCode(serverFactory *server.Factory) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		// Try to get the transaction data from the request body.
		var txn struct {
			Password  string `form:"password"`
			Password2 string `form:"password2"`
			UserID    string `form:"userId"`
			Code      string `form:"code"`
		}

		if err := ctx.Bind(&txn); err != nil {
			return derp.Wrap(err, "handler.PostResetCode", "Error binding form data")
		}

		// RULE: Ensure that passwords match
		if txn.Password != txn.Password2 {
			return derp.NewBadRequestError("handler.PostResetCode", "Passwords do not match")
		}

		// Try to get the factory for this domain
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.NewInternalError("handler.GetResetCode", "Invalid domain")
		}

		// Try to load the user by userID and resetCode
		userService := factory.User()

		user := model.NewUser()

		if err := userService.LoadByResetCode(txn.UserID, txn.Code, &user); err != nil {
			return derp.Wrap(err, "handler.GetResetCode", "Error loading user")
		}

		// Update the user with the new password
		user.SetPassword(txn.Password)

		if err := userService.Save(&user, "Updated Password"); err != nil {
			return derp.Wrap(err, "handler.GetResetCode", "Error saving user")
		}

		// Forward to the sign-in page with a success message
		return ctx.Redirect(http.StatusSeeOther, "/signin?message=password-reset")
	}
}
