package main

import "github.com/MRegterschot/GbxConnector/app"

func main() {
	// setup and run app
	err := app.SetupAndRunApp()
	if err != nil {
		panic(err)
	}
}
