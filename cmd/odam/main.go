package main

import (
	"flag"
	"log"

	"github.com/LdDl/odam"
)

func main() {
	settingsFile := flag.String("settings", "conf.json", "Path to application's settings")
	/* Read settings */
	flag.Parse()
	settings, err := odam.NewSettings(*settingsFile)
	if err != nil {
		log.Println(err)
		return
	}

	/* Initialize application */
	app, err := odam.NewApp(settings)
	if err != nil {
		log.Println(err)
		return
	}
	defer app.Close()

	err = app.Run()
	if err != nil {
		log.Println(err)
		return
	}

}
