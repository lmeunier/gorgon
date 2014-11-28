package main

import (
	"flag"
	"github.com/lmeunier/gorgon/app"
)

func main() {
	config_file := flag.String("c", "gorgon.ini", "Path to the Gorgon configuration file.")
	flag.Parse()

	gorgon_app := app.NewApp(*config_file)
	gorgon_app.Logger.Info("Starting Gorgon v" + app.Version + " (config: " + *config_file + ")")
	panic(gorgon_app.ListenAndServe())
}
