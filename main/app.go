package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ant0ine/go-json-rest/rest"
)

type App struct {
	router rest.App
	Config
}

type Config struct {
	Port     int `json:"port"`
	Database struct {
		Host     string `json:"host"`
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"database"`
}

func NewApp() (*App, error) {
	app := App{}

	router, err := rest.MakeRouter()
	if err != nil {
		return nil, err
	}

	err = app.loadConfig()
	if err != nil {
		return nil, err
	}

	app.router = router
	return &app, nil
}

func (app *App) loadConfig() error {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &app.Config)
}

func (app *App) AppFunc() rest.HandlerFunc {
	return app.router.AppFunc()
}
