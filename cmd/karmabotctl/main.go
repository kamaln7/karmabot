package main

import (
	"os"
	"time"

	"github.com/kamaln7/karmabot"
	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/ui/webui"

	"github.com/aybabtme/log"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli"
)

var (
	ll *log.Log
)

func main() {
	// logging

	ll = log.KV("version", karmabot.Version)

	// app
	app := cli.NewApp()
	app.Name = "karmabotctl"
	app.Version = karmabot.Version
	app.Usage = "manually manage karmabot"

	// general flags

	dbpath := cli.StringFlag{
		Name:  "dbpath",
		Value: "./db.sqlite3",
		Usage: "path to sqlite database",
	}

	debug := cli.BoolFlag{
		Name:  "debug",
		Usage: "set debug mode",
	}

	leaderboardlimit := cli.IntFlag{
		Name:  "leaderboardlimit",
		Value: 10,
		Usage: "the default amount of users to list in the leaderboard",
	}

	app.Commands = []cli.Command{
		{
			Name: "webui",
			Subcommands: []cli.Command{{
				Name:  "totp",
				Usage: "get a URL with a valid TOTP token",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "totp",
						Usage: "totp key",
					},
				},
				Action: mktotp,
			},
				{
					Name:  "serve",
					Usage: "start a webserver",
					Flags: []cli.Flag{
						dbpath,
						debug,
						leaderboardlimit,
						cli.StringFlag{
							Name:  "totp",
							Usage: "totp key",
						},
						cli.StringFlag{
							Name:  "path",
							Usage: "path to web UI files",
						},
						cli.StringFlag{
							Name:  "listenaddr",
							Usage: "address to listen and serve the web ui on",
						},
						cli.StringFlag{
							Name:  "url",
							Usage: "url address for accessing the web ui",
						},
					},
					Action: serve,
				}},
		},
	}

	app.Run(os.Args)
}

func getDB(path string) *database.DB {
	db, err := database.New(&database.Config{
		Path: path,
	})

	if err != nil {
		ll.KV("path", path).Err(err).Fatal("could not open sqlite db")
	}

	return db
}

func serve(c *cli.Context) error {
	db := getDB(c.String("dbpath"))

	ui, err := webui.New(&webui.Config{
		ListenAddr:       c.String("listenaddr"),
		URL:              c.String("url"),
		FilesPath:        c.String("path"),
		TOTP:             c.String("totp"),
		LeaderboardLimit: c.Int("leaderboardlimit"),
		Log:              ll.KV("provider", "webui"),
		Debug:            c.Bool("debug"),
		DB:               db,
	})

	if err != nil {
		ll.Err(err).Fatal("could not initialize web ui")
	}

	ui.Listen()
	return nil
}

func mktotp(c *cli.Context) error {
	TOTP := c.String("totp")
	token, err := totp.GenerateCode(TOTP, time.Now())
	if err != nil {
		ll.Err(err).Fatal("could not generate token")
	}

	ll.KV("token", token).Info("generated token")

	return nil
}
