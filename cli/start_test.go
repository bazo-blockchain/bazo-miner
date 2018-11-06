package cli

import (
	"bytes"
	cli2 "github.com/urfave/cli"
	"log"
	"testing"
)

func TestAddStartCommand(t *testing.T) {
	dummyLogger := log.New(&bytes.Buffer{}, "", 0)
	app := cli2.NewApp()
	AddStartCommand(app, dummyLogger)

	command := app.Command("start")
	if command == nil {
		t.Errorf("Did not add start command to cli app.")
	}
}