package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/VannRR/robuku/bukudb"
	"github.com/VannRR/robuku/inputhandler"
	rofiapi "github.com/VannRR/rofi-api"
)

const (
	bukuDbEnvVar      = "ROBUKU_DB_PATH"
	xdgDataHomeEnvVar = "XDG_DATA_HOME"
)

func main() {
	api, err := rofiapi.NewRofiApi(inputhandler.Data{})
	handleInitError(api, err)
	if api.Data.State != inputhandler.StateErrorSelect {
		defer api.Draw()
	}

	bukuDbPath, err := getBukuDbPath()
	if err != nil {
		inputhandler.SetMessageToError(api, err)
		return
	}

	db, err := bukudb.NewBukuDB(bukuDbPath)
	if err != nil {
		inputhandler.SetMessageToError(api, err)
		return
	}

	in := inputhandler.NewInputHandler(db, api)
	handleApiInput(api, in)
}

func handleInitError(api *rofiapi.RofiApi[inputhandler.Data], err error) {
	if !api.IsRanByRofi() {
		fmt.Println("this is a rofi script, for more information check the rofi manual")
	}

	if api.Data.State == inputhandler.StateErrorShow {
		api.Data.State = inputhandler.StateErrorSelect
	}

	if err != nil {
		inputhandler.SetMessageToError(api, err)
	}
}

func getBukuDbPath() (string, error) {
	if path := os.Getenv(bukuDbEnvVar); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if xdgDataHomeDir := os.Getenv(xdgDataHomeEnvVar); xdgDataHomeDir != "" {
		path := filepath.Join(xdgDataHomeDir, "buku/bookmarks.db")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if homeDir, _ := os.UserHomeDir(); homeDir != "" {
		path := filepath.Join(homeDir, ".local/share/buku/bookmarks.db")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf(
		"could not find buku bookmarks db path, try setting the env variable $%s",
		bukuDbEnvVar)
}

func handleApiInput(api *rofiapi.RofiApi[inputhandler.Data], in *inputhandler.InputHandler) {
	if selected, ok := api.GetSelectedEntry(); ok {
		in.HandleInput(selected.Text)
	} else {
		in.HandleBookmarksShow()
	}
}
