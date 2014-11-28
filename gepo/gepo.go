package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "gepo"
	app.Usage = "Geppetto command line interface."
	app.Action = func(c *cli.Context) {
	}

	app.Run(os.Args)
}
