package main

import (
	"github.com/lmeunier/gorgon/app"
	"net/http"
)

const (
	ADDRESS = ":5000"
)

func main() {
	app := app.NewApp()
	http.Handle("/", app.Router)
	http.ListenAndServe(ADDRESS, nil)
}
