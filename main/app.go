package main

import (
	"github.com/ant0ine/go-json-rest/rest"
)

type App struct {
	router rest.App
}

func NewApp() (*App, error) {
	app := App{}

	router, err := rest.MakeRouter()
	if err != nil {
		return nil, err
	}

	app.router = router
	return &app, nil
}

func (app *App) AppFunc() rest.HandlerFunc {
	return app.router.AppFunc()
}
