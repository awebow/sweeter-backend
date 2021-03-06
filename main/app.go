package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type App struct {
	router rest.App
	Config
	DB *sqlx.DB
}

type Config struct {
	Port       int    `json:"port"`
	SigningKey string `json:"signing_key"`
	Database   struct {
		Host     string `json:"host"`
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"database"`
}

func NewApp() (*App, error) {
	app := App{}

	router, err := rest.MakeRouter(
		rest.Get("/users", app.GetUsers),
		rest.Get("/users/me", app.GetUsersMe),
		rest.Get("/users/me/sweets", app.GetUsersMeSweets),
		rest.Get("/users/me/newsfeeds", app.GetUsersMeNewsfeeds),
		rest.Get("/users/me/followings", app.GetUsersMeFollowings),
		rest.Get("/users/me/followers", app.GetUsersMeFollowers),
		rest.Post("/users/me/followings/:no", app.PostUsersMeFollowings),
		rest.Delete("/users/me/followings/:no", app.DeleteUsersMeFollowings),
		rest.Get("/users/me/relations/:no", app.GetUsersRelations),
		rest.Put("/users/me/picture", app.PutUsersMePicture),
		rest.Get("/users/:no", app.GetUsers),
		rest.Get("/users/:no/sweets", app.GetUsersSweets),
		rest.Get("/users/:no/followings", app.GetUsersFollowings),
		rest.Get("/users/:no/followers", app.GetUsersFollowers),
		rest.Post("/users", app.PostUsers),
		rest.Post("/users/tokens", app.PostTokens),

		rest.Get("/sweets/:no", app.GetSweets),
		rest.Post("/sweets", app.PostSweets),
	)
	if err != nil {
		return nil, err
	}

	err = app.loadConfig()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", app.Database.User, app.Database.Password, app.Database.Host, app.Database.Name))
	if err != nil {
		return nil, err
	}

	app.DB = db.Unsafe()

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
