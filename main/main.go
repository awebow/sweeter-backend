package main

import (
	"fmt"
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
	fmt.Printf("Sweeter is starting on port %d\n...", app.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.Port), api.MakeHandler()))
}
