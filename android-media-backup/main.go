package main

import (
	"log"
	"os"

	"gioui.org/app"
	_ "gioui.org/app/permission/storage"
	"github.com/psanford/android-media-backup-go-experiment/ui"
)

func main() {
	gui := ui.New()
	go func() {
		if err := gui.Run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
