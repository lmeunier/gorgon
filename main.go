package main

import (
	"github.com/lmeunier/gorgon/app"
)

func main() {
	app := app.NewApp("gorgon.ini")
	panic(app.ListenAndServe())
}
