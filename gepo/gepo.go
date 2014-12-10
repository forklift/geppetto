package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/forklift/geppetto/api"
)

var (
	server *api.Server
	Log    *logrus.Logger
)

func main() {

	app := cli.NewApp()
	app.Name = "gepo"
	app.Usage = "Geppetto command line interface."
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Be talkative.",
		},
		cli.BoolFlag{
			Name:  "robot",
			Usage: "More structure and parsable output.",
		},
		cli.StringFlag{
			Name:   "endpoint",
			Value:  "http://127.0.0.1:5000",
			Usage:  "The Geppetto endpoint.",
			EnvVar: "GEPPETOO_ENDPOINT",
		},
	}

	app.Action = func(c *cli.Context) {
		cli.ShowSubcommandHelp(c)
	}

	app.Commands = []cli.Command{
		ping,
	}

	app.Before = func(c *cli.Context) error {

		var err error

		Log = logrus.New()

		server, err = api.NewClient(c.String("endpoint"))

		if err == nil {
			err = server.Ping()
		}

		if err != nil {
			Log.Error(err)
			return err
		}
		return nil
	}
	app.Run(os.Args)
}

var ping = cli.Command{
	Name:  "ping",
	Usage: "Ping the server.",
	Action: func(c *cli.Context) {
		err := server.Ping()
		if err != nil {
			Log.Fatal(err)
		}
		Log.Info("PONG")
	},
}
