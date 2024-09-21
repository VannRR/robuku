package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/VannRR/robuku/bukudb"
	"github.com/VannRR/robuku/inputhandler"
	"github.com/VannRR/robuku/rofidata"
	"github.com/VannRR/rofi-api"
)

const (
	bukuDbEnvVar      = "ROBUKU_DB_PATH"
	xdgDataHomeEnvVar = "XDG_DATA_HOME"
)

func main() {
	data := rofidata.Data{}
	api, err := rofiapi.NewRofiApi(&data)
	if err != nil {
		inputhandler.SetMessageToError(api, err)
		return
	}
	bukuDbPath := getBukuDbPath()

	if bukuDbPath == "" {
		handleMissingDbPath(api)
		return
	}

	db, err := bukudb.NewBukuDB(bukuDbPath)
	if err != nil {
		inputhandler.SetMessageToError(api, err)
		return
	}

	in := inputhandler.NewInputHandler(db, api)
	handleArgs(in)
	api.Draw()
}

func getBukuDbPath() string {
	if path := os.Getenv(bukuDbEnvVar); path != "" {
		return path
	}

	if xdgDataHomeDir := os.Getenv(xdgDataHomeEnvVar); xdgDataHomeDir != "" {
		return filepath.Join(xdgDataHomeDir, "buku/bookmarks.db")
	}

	if homeDir, _ := os.UserHomeDir(); homeDir != "" {
		return filepath.Join(homeDir, ".local/share/buku/bookmarks.db")
	}

	return ""
}

func handleMissingDbPath(api *rofiapi.RofiApi[*rofidata.Data]) {
	xdg := "$XDG_DATA_HOME/buku/bookmarks.db"
	home := "$HOME/.local/share/buku/bookmarks.db"
	err := fmt.Errorf(
		"could not find path %s or %s, try setting the env variable $%s",
		xdg, home, bukuDbEnvVar)
	inputhandler.SetMessageToError(api, err)
}

func handleArgs(in *inputhandler.InputHandler) {
	if len(os.Args) <= 1 {
		in.HandleBookmarksShow()
	} else {
		in.HandleInput(os.Args[1])
	}
}
