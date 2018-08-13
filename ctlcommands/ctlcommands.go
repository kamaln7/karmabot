package ctlcommands

import (
	"fmt"
	"time"

	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/ui/webui"

	"github.com/aybabtme/log"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli"
)

type Commands struct {
	Logger *log.Log
}

func (cc *Commands) Serve(c *cli.Context) error {
	db := cc.getDB(c.String("db"))
	TOTP := c.String("totp")

	ui, err := webui.New(&webui.Config{
		ListenAddr:       c.String("listenaddr"),
		URL:              c.String("url"),
		FilesPath:        c.String("path"),
		TOTP:             TOTP,
		LeaderboardLimit: c.Int("leaderboardlimit"),
		Log:              cc.Logger.KV("provider", "webui"),
		Debug:            c.Bool("debug"),
		DB:               db,
	})

	if err != nil {
		cc.Logger.Err(err).Fatal("could not initialize web ui")
		return err
	}

	token, err := totp.GenerateCode(TOTP, time.Now())
	if err != nil {
		cc.Logger.Err(err).Fatal("could not generate totp token")
	} else {
		cc.Logger.KV("token", token).Info("generated totp token")
	}

	ui.Listen()
	return nil
}

func (cc *Commands) Mktotp(c *cli.Context) error {
	TOTP := c.String("totp")
	token, err := totp.GenerateCode(TOTP, time.Now())
	if err != nil {
		cc.Logger.Err(err).Fatal("could not generate token")
	}

	cc.Logger.KV("token", token).Info("generated token")

	return nil
}

func (cc *Commands) AddKarma(c *cli.Context) error {
	var (
		db     = cc.getDB(c.String("db"))
		from   = c.String("from")
		to     = c.String("to")
		reason = c.String("reason")
		points = c.Int("points")
	)

	if from == "" || to == "" {
		cc.Logger.Fatal("please pass valid users to the `to` and `from` options")
	}

	if points == 0 {
		cc.Logger.Fatal("you may not add 0 points to a user")
	}

	record := &database.Points{
		From:   from,
		To:     to,
		Reason: reason,
		Points: points,
	}

	err := db.InsertPoints(record)
	if err != nil {
		cc.Logger.Err(err).Fatal("could not insert record")
	}

	cc.Logger.Info("inserted record")

	return nil
}

func (cc *Commands) MigrateKarma(c *cli.Context) error {
	var (
		db   = cc.getDB(c.String("db"))
		from = c.String("from")
		to   = c.String("to")
	)

	if from == "" || to == "" {
		cc.Logger.Fatal("please pass valid users to the `to` and `from` options")
	}

	user, err := db.GetUser(from)
	if err != nil {
		cc.Logger.Err(err).KV("from", from).Fatal("could not look up user `from`")
	}

	if user.Points == 0 {
		cc.Logger.KV("from", from).Fatal("user does not have any points")
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
			cc.Logger.Err(err).Fatal("could not insert record")
		}
	}

	cc.Logger.KV("from", from).KV("to", to).KV("points", user.Points).Info("migrated karma")

	return nil
}

func (cc *Commands) ResetKarma(c *cli.Context) error {
	var (
		db   = cc.getDB(c.String("db"))
		name = c.String("user")
	)

	if name == "" {
		cc.Logger.Fatal("please pass a valid user to the `user` option")
	}

	user, err := db.GetUser(name)
	if err != nil {
		cc.Logger.Err(err).KV("user", name).Fatal("could not look up user")
	}

	err = db.InsertPoints(&database.Points{
		From:   "karmabot",
		To:     name,
		Points: -1 * user.Points,
		Reason: "karmabotctl resetting karma",
	})

	if err != nil {
		cc.Logger.Err(err).Fatal("could not insert record")
	}

	cc.Logger.KV("user", name).Info("reset karma")

	return nil
}

func (cc *Commands) SetKarma(c *cli.Context) error {
	var (
		db     = cc.getDB(c.String("db"))
		name   = c.String("user")
		points = c.Int("points")
	)

	if name == "" {
		cc.Logger.Fatal("please pass a valid user to the `user` option")
	}

	user, err := db.GetUser(name)
	if err != nil {
		cc.Logger.Err(err).KV("user", name).Fatal("could not look up user")
	}

	err = db.InsertPoints(&database.Points{
		From:   "karmabot",
		To:     name,
		Points: points - user.Points,
		Reason: "karmabotctl overriding karma",
	})

	if err != nil {
		cc.Logger.Err(err).Fatal("could not insert record")
	}

	cc.Logger.KV("user", name).KV("points", points).Info("set karma")

	return nil
}

func (cc *Commands) GetThrowback(c *cli.Context) error {
	var (
		user = c.String("user")
		db   = cc.getDB(c.String("db"))
	)

	if user == "" {
		cc.Logger.Fatal("please pass a valid user to the `user` option")
	}

	throwback, err := db.GetThrowback(user)
	if err != nil {
		cc.Logger.Err(err).Fatal("could not look up user data")
	}

	cc.Logger.KV("throwback", throwback).Info("got throwback")
	return nil
}

func (cc *Commands) getDB(path string) *database.DB {
	db, err := database.New(&database.Config{
		Path: path,
	})

	if err != nil {
		cc.Logger.KV("path", path).Err(err).Fatal("could not open sqlite db")
	}

	return db
}
