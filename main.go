package main

import (
	"flag"
	"github.com/lmeunier/gorgon/app"
	"os"
)

func main() {
	config_file := flag.String("c", "gorgon.ini", "Path to the Gorgon configuration file.")
	version := flag.Bool("v", false, "Prints the Gorgon version and exits.")
	flag.Parse()

	if *version {
		// prints the Gorgon version and exits
		print("Gorgon v" + app.Version + "\n")
		os.Exit(1)
	}

	gorgon_app := app.NewApp(*config_file)
	gorgon_app.Logger.Info("Starting Gorgon v" + app.Version + " (config: " + *config_file + ")")
	panic(gorgon_app.ListenAndServe())
}
