package karmabot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/munge"
	"github.com/kamaln7/karmabot/ui"

	"github.com/aybabtme/log"
	"github.com/nlopes/slack"
)

var (
	regexps = struct {
		Motivate, GiveKarma, Leaderboard, URL, SlackUser *regexp.Regexp
	}{
		Motivate:    karmaReg.GetMotivate(),
		GiveKarma:   karmaReg.GetGive(),
		Leaderboard: regexp.MustCompile(`^karma(?:bot)? (?:leaderboard|top|highscores) ?([0-9]+)?$`),
		URL:         regexp.MustCompile(`^karma(?:bot)? (?:url|web|link)?$`),
		SlackUser:   regexp.MustCompile(`^<@([A-Za-z0-9]+)>$`),
	}
)

// Slack contains the Slack client and RTM object.
type Slack struct {
	Bot *slack.Client
	RTM *slack.RTM
}

// Config contains all the necessary configs for karmabot.
type Config struct {
	Slack                       *Slack
	Debug, Motivate             bool
	MaxPoints, LeaderboardLimit int
	Log                         *log.Log
	UI                          ui.Provider
	DB                          *database.DB
}

// A Bot is an instance of karmabot.
type Bot struct {
	Config *Config
}

// New returns a pointer to an new instance of karmabot.
func New(config *Config) *Bot {
	return &Bot{
		Config: config,
	}
}

// Listen starts listening for Slack messages and calls the
// appropriate handlers.
func (b *Bot) Listen() {
	for {
		select {
		case msg := <-b.Config.Slack.RTM.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				go b.handleMessageEvent(msg.Data.(*slack.MessageEvent))
			case *slack.ConnectedEvent:
				b.Config.Log.Info("connected to slack")

				if b.Config.Debug {
					b.Config.Log.KV("info", ev.Info).Info("got slack info")
					b.Config.Log.KV("connections", ev.ConnectionCount).Info("got connection count")
				}
			case *slack.RTMError:
				b.Config.Log.Err(ev).Error("slack rtm error")
			case *slack.InvalidAuthEvent:
				b.Config.Log.Fatal("invalid slack token")
			default:
			}
		}
	}
}

// SendMessage sends a message to a Slack channel.
func (b *Bot) SendMessage(message, channel string) {
	b.Config.Slack.RTM.SendMessage(b.Config.Slack.RTM.NewOutgoingMessage(message, channel))
}

func (b *Bot) handleError(err error, channel string) bool {
	if err == nil {
		return false
	}

	b.Config.Log.Err(err).Error("error")
	var message string
	if b.Config.Debug {
		message = err.Error()
	} else {
		message = "an error has occurred."
	}

	b.SendMessage(message, channel)
	return true
}

func (b *Bot) handleMessageEvent(ev *slack.MessageEvent) {
	if ev.Type != "message" {
		return
	}

	// convert motivates into karmabot syntax
	if b.Config.Motivate {
		if match := regexps.Motivate.FindStringSubmatch(ev.Text); len(match) > 0 {
			ev.Text = match[1] + "++ for doing good work"
		}
	}

	switch {
	case regexps.URL.MatchString(ev.Text):
		b.printURL(ev)

	case regexps.GiveKarma.MatchString(ev.Text):
		b.givePoints(ev)

	case regexps.Leaderboard.MatchString(ev.Text):
		b.printLeaderboard(ev)
	}
}

func (b *Bot) printURL(ev *slack.MessageEvent) {
	url, err := b.Config.UI.GetURL("/")
	if b.handleError(err, ev.Channel) {
		return
	}

	// ui is disabled
	if url == "" {
		return
	}

	b.SendMessage(url, ev.Channel)
}

func (b *Bot) givePoints(ev *slack.MessageEvent) {
	match := regexps.GiveKarma.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		return
	}

	// forgive me
	if match[1] != "" {
		// we matched the first alt expression
		match = match[:4]
	} else {
		// we matched the second alt expression
		match = append(match[:1], match[4:]...)
	}

	from, err := b.getUserNameByID(ev.User)
	if b.handleError(err, ev.Channel) {
		return
	}
	to, err := b.parseUser(match[1])
	if b.handleError(err, ev.Channel) {
		return
	}
	to = strings.ToLower(to)

	points := min(len(match[2])-1, b.Config.MaxPoints)
	if match[2][0] == '-' {
		points *= -1
	}
	reason := match[3]

	record := &database.Points{
		From:   from,
		To:     to,
		Points: points,
		Reason: reason,
	}

	err = b.Config.DB.InsertPoints(record)
	if b.handleError(err, ev.Channel) {
		return
	}

	user, err := b.Config.DB.GetUser(to)
	if b.handleError(err, ev.Channel) {
		return
	}

	text := fmt.Sprintf("%s == %d (", to, user.Points)

	if points > 0 {
		text += "+"
	}
	text = fmt.Sprintf("%s%d", text, points)

	if reason != "" {
		text += fmt.Sprintf(" for %s", reason)
	}
	text += ")"

	b.SendMessage(text, ev.Channel)
}

func (b *Bot) printLeaderboard(ev *slack.MessageEvent) {
	match := regexps.Leaderboard.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		return
	}

	limit := b.Config.LeaderboardLimit
	if match[1] != "" {
		var err error
		limit, err = strconv.Atoi(match[1])
		if b.handleError(err, ev.Channel) {
			return
		}
	}

	text := fmt.Sprintf("*top %d leaderboard*\n", limit)

	url, err := b.Config.UI.GetURL(fmt.Sprintf("/leaderboard/%d", limit))
	if b.handleError(err, ev.Channel) {
		return
	}
	if url != "" {
		text = fmt.Sprintf("%s%s\n", text, url)
	}

	leaderboard, err := b.Config.DB.GetLeaderboard(limit)
	if b.handleError(err, ev.Channel) {
		return
	}

	for i, user := range leaderboard {
		text += fmt.Sprintf("%d. %s == %d\n", i+1, munge.Munge(user.Name), user.Points)
	}

	b.SendMessage(text, ev.Channel)
}

func (b *Bot) parseUser(user string) (string, error) {
	if match := regexps.SlackUser.FindStringSubmatch(user); len(match) > 0 {
		return b.getUserNameByID(match[1])
	}

	return user, nil
}

func (b *Bot) getUserNameByID(id string) (string, error) {
	userInfo, err := b.Config.Slack.Bot.GetUserInfo(id)
	if err != nil {
		return "", err
	}

	return userInfo.Name, nil
}
