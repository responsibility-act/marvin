package app

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	accountmiddleware "github.com/marvin-automator/marvin/accounts/middleware"

	"context"
	accounthandlers "github.com/marvin-automator/marvin/accounts/handlers"
	actionhandlers "github.com/marvin-automator/marvin/actions/handlers"
	"github.com/marvin-automator/marvin/handlers"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// T is a translator.
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.Automatic(buffalo.Options{
			Env:         ENV,
			SessionName: "marvin_session",
			Context:     context.Background(),
		})

		// Automatically save sessions
		app.Use(middleware.SessionSaver)

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		if ENV != "test" {
			// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
			// Remove to disable this.
			app.Use(csrf.Middleware)
		}

		// Setup and use translations:
		var err error
		if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
			err = app.Stop(err)
			if err != nil {
				panic(err)
			}
		}

		app.Use(T.Middleware())

		app.Redirect(302, "/", "/dashboard")
		app.GET("/login", bf(accounthandlers.LoginPage))
		app.POST("/login", bf(accounthandlers.ProcessLogin))

		g := app.Group("/app")
		g.Use(accountmiddleware.Middleware)

		// API
		//  Chores
		g.GET("/chores", bf(actionhandlers.AccountChores))
		//  Actions
		g.GET("/actions", bf(actionhandlers.ActionGroups))

		app.ServeFiles("/assets", packr.NewBox("./public/assets"))
	}

	return app
}

// bf turns a marvin handler into a buffalo handler
func bf(h handlers.Handler) buffalo.Handler {
	return h.ToBuffalo()
}
