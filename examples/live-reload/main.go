package main

import (
	"log"
	"os"
	fp "path/filepath"

	"github.com/jamalsa/qamel"
)

func main() {
	// Create application
	app := qamel.NewApplication(len(os.Args), os.Args)
	app.SetApplicationDisplayName("Live Reload Example")

	// Create a QML viewer
	view := qamel.NewViewerWithSource("res/main.qml")
	view.SetResizeMode(qamel.SizeRootObjectToView)
	view.SetHeight(300)
	view.SetWidth(400)
	view.Show()

	// Watch change in resource dir
	projectDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Failed to get working directory:", err)
	}

	resDir := fp.Join(projectDir, "res")
	go view.WatchResourceDir(resDir)

	// Exec app
	app.Exec()
}
