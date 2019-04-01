package main

import (
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

func main() {
	api := rest.NewApi()
	app, err := NewApp()

	if err != nil {
		log.Fatal(err)
		return
	}

	api.SetApp(app)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
