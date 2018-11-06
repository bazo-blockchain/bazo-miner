package cli

import (
	cli2 "github.com/urfave/cli"
	"testing"
)

func TestAddGenerateCommand(t *testing.T) {
	app := cli2.NewApp()
	AddGenerateCommand(app)

	command := app.Command("generate")
	if command == nil {
		t.Errorf("Did not add generate command to cli app.")
	}
}