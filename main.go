package main

import (
	"flag"
	"github.com/lmeunier/gorgon/app"
)

func main() {
	config_file := flag.String("c", "gorgon.ini", "Path to the Gorgon configuration file.")
	flag.Parse()

	app := app.NewApp(*config_file)
	panic(app.ListenAndServe())
}
