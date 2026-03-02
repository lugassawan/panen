package main

import (
	"embed"

	"github.com/lugassawan/panen/backend/presenter"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	a := presenter.NewApp()

	err := wails.Run(&options.App{
		Title:  "Panen",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: a.Startup,
		Bind: []any{
			a,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
