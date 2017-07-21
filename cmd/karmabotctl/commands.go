package main

import (
	"time"

	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/ui/webui"

	"fmt"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli"
)

func serve(c *cli.Context) error {
	db := getDB(c.String("db"))
	TOTP := c.String("totp")

	ui, err := webui.New(&webui.Config{
		ListenAddr:       c.String("listenaddr"),
		URL:              c.String("url"),
		FilesPath:        c.String("path"),
		TOTP:             TOTP,
		LeaderboardLimit: c.Int("leaderboardlimit"),
		Log:              ll.KV("provider", "webui"),
		Debug:            c.Bool("debug"),
		DB:               db,
	})

	if err != nil {
		ll.Err(err).Fatal("could not initialize web ui")
		return err
	}

	token, err := totp.GenerateCode(TOTP, time.Now())
	if err != nil {
		ll.Err(err).Fatal("could not generate totp token")
	} else {
		ll.KV("token", token).Info("generated totp token")
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

func addKarma(c *cli.Context) error {
	var (
		db     = getDB(c.String("db"))
		from   = c.String("from")
		to     = c.String("to")
		reason = c.String("reason")
		points = c.Int("points")
	)

	if from == "" || to == "" {
		ll.Fatal("please pass valid users to the `to` and `from` options")
	}

	if points == 0 {
		ll.Fatal("you may not add 0 points to a user")
	}

	record := &database.Points{
		From:   from,
		To:     to,
		Reason: reason,
		Points: points,
	}

	err := db.InsertPoints(record)
	if err != nil {
		ll.Err(err).Fatal("could not insert record")
	}

	ll.Info("inserted record")

	return nil
}

func migrateKarma(c *cli.Context) error {
	var (
		db   = getDB(c.String("db"))
		from = c.String("from")
		to   = c.String("to")
	)

	if from == "" || to == "" {
		ll.Fatal("please pass valid users to the `to` and `from` options")
	}

	user, err := db.GetUser(from)
	if err != nil {
		ll.Err(err).KV("from", from).Fatal("could not look up user `from`")
	}

	if user.Points == 0 {
		ll.KV("from", from).Fatal("user does not have any points")
	}

	reason := fmt.Sprintf("migrating karma from %s to %s", from, to)
	records := []*database.Points{
		// remove points from `from`
		{
			From:   "karmabot",
			To:     from,
			Reason: reason,
			Points: -user.Points,
		},
		// add points to `to`
		{
			From:   "karmabot",
			To:     to,
			Reason: reason,
			Points: user.Points,
		},
	}

	for _, record := range records {
		err := db.InsertPoints(record)
		if err != nil {
			ll.Err(err).Fatal("could not insert record")
		}
	}

	ll.KV("from", from).KV("to", to).KV("points", user.Points).Info("migrated karma")

	return nil
}

func resetKarma(c *cli.Context) error {
	var (
		db   = getDB(c.String("db"))
		name = c.String("user")
	)

	if name == "" {
		ll.Fatal("please pass a valid user to the `user` option")
	}

	user, err := db.GetUser(name)
	if err != nil {
		ll.Err(err).KV("user", name).Fatal("could not look up user")
	}

	err = db.InsertPoints(&database.Points{
		From:   "karmabot",
		To:     name,
		Points: -1 * user.Points,
		Reason: "karmabotctl resetting karma",
	})

	if err != nil {
		ll.Err(err).Fatal("could not insert record")
	}

	ll.KV("user", name).Info("reset karma")

	return nil
}

func setKarma(c *cli.Context) error {
	var (
		db     = getDB(c.String("db"))
		name   = c.String("user")
		points = c.Int("points")
	)

	if name == "" {
		ll.Fatal("please pass a valid user to the `user` option")
	}

	user, err := db.GetUser(name)
	if err != nil {
		ll.Err(err).KV("user", name).Fatal("could not look up user")
	}

	err = db.InsertPoints(&database.Points{
		From:   "karmabot",
		To:     name,
		Points: points - user.Points,
		Reason: "karmabotctl overriding karma",
	})

	if err != nil {
		ll.Err(err).Fatal("could not insert record")
	}

	ll.KV("user", name).KV("points", points).Info("set karma")

	return nil
}
