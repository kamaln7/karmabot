package main

import (
	"flag"

	"github.com/kamaln7/karmabot"
	"github.com/kamaln7/karmabot/database"
	karmabotui "github.com/kamaln7/karmabot/ui"
	blankui "github.com/kamaln7/karmabot/ui/providers/blank"
	webui "github.com/kamaln7/karmabot/ui/providers/web"

	"github.com/aybabtme/log"
	"github.com/nlopes/slack"
)

func main() {
	// startup

	ll := log.KV("version", karmabot.VERSION)
	ll.Info("starting karmabot")

	// config

	flags := struct {
		Token, DB, WebUITOTP, WebUIPath, WebUIListenAddr, WebUIURL *string
		MaxPoints, LeaderboardLimit                                *int
		Debug, Motivate                                            *bool
	}{
		Token:            flag.String("token", "", "slack RTM token"),
		DB:               flag.String("db", "./db.sqlite3", "path to sqlite database"),
		MaxPoints:        flag.Int("maxpoints", 6, "the maximum amount of points that users can give/take at once"),
		LeaderboardLimit: flag.Int("leaderboardlimit", 10, "the default amount of users to list in the leaderboard"),
		Debug:            flag.Bool("debug", false, "set debug mode"),
		WebUITOTP:        flag.String("webui.totp", "", "totp key"),
		WebUIPath:        flag.String("webui.path", "", "path to web UI files"),
		WebUIListenAddr:  flag.String("webui.listenaddr", "", "address to listen and serve the web ui on"),
		WebUIURL:         flag.String("webui.url", "", "url address for accessing the web ui"),
		Motivate:         flag.Bool("motivate", true, "toggle motivate.im support"),
	}
	flag.Parse()

	// database

	DB, err := database.New(&database.Config{
		Path: *flags.DB,
	})

	if err != nil {
		ll.KV("path", *flags.DB).Err(err).Fatal("could not open sqlite db")
	}

	// slack

	if *flags.Token == "" {
		ll.Fatal("please pass the slack RTM token (see `karmabot -h` for help)")
	}

	//TODO: figure out a way to fix this
	//our current logging library does not implement
	//log.Logger
	//slack.SetLogger(*ll)
	sc := &karmabot.Slack{
		Bot: slack.New(*flags.Token),
	}
	sc.Bot.SetDebug(*flags.Debug)
	sc.RTM = sc.Bot.NewRTM()

	go sc.RTM.ManageConnection()

	// karmabot

	var ui karmabotui.Provider
	if *flags.WebUIPath != "" && *flags.WebUIListenAddr != "" {
		ui, err = webui.New(&webui.Config{
			ListenAddr:       *flags.WebUIListenAddr,
			URL:              *flags.WebUIURL,
			FilesPath:        *flags.WebUIPath,
			TOTP:             *flags.WebUITOTP,
			LeaderboardLimit: *flags.LeaderboardLimit,
			Log:              ll.KV("provider", "webui"),
			Debug:            *flags.Debug,
		})

		if err != nil {
			ll.Err(err).Fatal("could not initialize web ui")
		}
	} else {
		ui = blankui.New()
	}
	go ui.Listen()

	bot := karmabot.New(&karmabot.Config{
		Slack:            sc,
		UI:               ui,
		Debug:            *flags.Debug,
		MaxPoints:        *flags.MaxPoints,
		LeaderboardLimit: *flags.LeaderboardLimit,
		Log:              ll,
		DB:               DB,
	})

	bot.Listen()
}
