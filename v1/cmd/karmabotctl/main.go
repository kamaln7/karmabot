package main

import (
	"os"

	"github.com/kamaln7/karmabot"
	"github.com/kamaln7/karmabot/ctlcommands"

	"github.com/aybabtme/log"
	"github.com/urfave/cli"
)

var (
	ll *log.Log
)

func main() {
	// logging

	ll = log.KV("version", karmabot.Version)

	// commands

	cc := &ctlcommands.Commands{
		Logger: ll,
	}

	// app
	app := cli.NewApp()
	app.Name = "karmabotctl"
	app.Version = karmabot.Version
	app.Usage = "manually manage karmabot"

	// general flags

	dbpath := cli.StringFlag{
		Name:  "db",
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

	// webui

	webuiCommands := []cli.Command{
		{
			Name:  "totp",
			Usage: "generate a TOTP token",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "totp",
					Usage: "totp key",
				},
			},
			Action: cc.Mktotp,
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
			Action: cc.Serve,
		},
	}

	// karma

	karmaCommands := []cli.Command{
		{
			Name:  "add",
			Usage: "add karma to a user",
			Flags: []cli.Flag{
				dbpath,
				cli.StringFlag{
					Name: "from",
				},
				cli.StringFlag{
					Name: "to",
				},
				cli.StringFlag{
					Name: "reason",
				},
				cli.IntFlag{
					Name: "points",
				},
			},
			Action: cc.AddKarma,
		},
		{
			Name:  "migrate",
			Usage: "move a user's karma to another user",
			Flags: []cli.Flag{
				dbpath,
				cli.StringFlag{
					Name: "from",
				},
				cli.StringFlag{
					Name: "to",
				},
				cli.StringFlag{
					Name: "reason",
				},
			},
			Action: cc.MigrateKarma,
		},
		{
			Name:  "reset",
			Usage: "reset a user's karma",
			Flags: []cli.Flag{
				dbpath,
				cli.StringFlag{
					Name: "user",
				},
			},
			Action: cc.ResetKarma,
		},
		{
			Name:  "set",
			Usage: "set a user's karma to a specific number",
			Flags: []cli.Flag{
				dbpath,
				cli.StringFlag{
					Name: "user",
				},
				cli.IntFlag{
					Name: "points",
				},
			},
			Action: cc.SetKarma,
		},
		{
			Name:  "throwback",
			Usage: "get a karma throwback for a user",
			Flags: []cli.Flag{
				dbpath,
				cli.StringFlag{
					Name: "user",
				},
			},
			Action: cc.GetThrowback,
		},
	}

	// main app

	app.Commands = []cli.Command{
		{
			Name:        "karma",
			Subcommands: karmaCommands,
		},
		{
			Name:        "webui",
			Subcommands: webuiCommands,
		},
	}

	app.Run(os.Args)
}
