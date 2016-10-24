package karmabot

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/math"
	"github.com/kamaln7/karmabot/ui"

	"github.com/nlopes/slack"
)

var (
	regexps = struct {
		Motivate, Karma, Leaderboard, URL, SlackUser *regexp.Regexp
	}{
		Motivate:    regexp.MustCompile(`^(?:\?|!)m\s+@?([^\s]+?)(?:\:\s)?$`),
		Karma:       regexp.MustCompile(`(?:^|\s+)@?([^\s]+?)\:?\s?([\+]{2,}|[\-]{2,})((?: for)? (.*))?$`),
		Leaderboard: regexp.MustCompile(`^karma(?:bot)? (?:leaderboard|top|highscores) ?([0-9]+)?$`),
		URL:         regexp.MustCompile(`^karma(?:bot)? (?:url|web|link)?$`),
		SlackUser:   regexp.MustCompile(`^<@([A-Za-z0-9]+)>$`),
	}
)

type Slack struct {
	Bot *slack.Client
	RTM *slack.RTM
}

type Config struct {
	Slack                       *Slack
	Debug                       bool
	MaxPoints, LeaderboardLimit int
	Log                         *log.Logger
	UI                          *ui.Provider
	DB                          *database.DB
}

type Bot struct {
	Config *Config
}

func New(config *Config) {
	return &Bot{
		Config: config,
	}
}

func (b *Bot) Listen() {
	for {
		select {
		case msg := <-b.Slack.RTM.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				go b.handleMessageEvent(msg.Data.(*slack.MessageEvent))
			case *slack.ConnectedEvent:
				ll.Info("connected to slack")

				if b.Config.Debug {
					ll.KV("info", ev.Info).Info("got slack info")
					ll.KV("connections", ev.ConnectionCount).Info("got connection count")
				}
			case *slack.RTMError:
				ll.Err(ev).Error("slack rtm error")
			case *slack.InvalidAuthEvent:
				ll.Fatal("invalid slack token")
			default:
			}
		}
	}
}

func (b *Bot) SendMessage(message, channel string) {
	return rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
}

func (b *Bot) handleError(err error, channel string) {
	if err == nil {
		return false
	}

	ll.Err(err).Error("error")
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

	case regexps.Karma.MatchString(ev.Text):
		b.givePoints(ev)

	case regexps.Leaderboard.MatchString(ev.Text):
		b.printLeaderboard(ev)
	}
}

func (b *Bot) printURL(ev *slack.MessageEvent) {
	if !hasWebUI {
		rtm.SendMessage(rtm.NewOutgoingMessage("webui not enabled. please pass the `-webuipath`, `-webuiurl`, and `-listenaddr` options in order to enable the web ui", ev.Channel))
		return
	}

	URL, err := b.UI.GetURL("/")
	if b.handleError(err, ev.Channel) {
		return
	}

	b.SendMessage(URL, ev.Channel)
}

func (b *Bot) givePoints(ev *slack.MessageEvent) {
	match := regexps.Karma.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		return
	}

	from, err := b.getUserNameByID(ev.User)
	if handleError(err, ev.Channel) {
		return
	}
	to, err := b.parseUser(match[1])
	if handleError(err, ev.Channel) {
		return
	}
	to = strings.ToLower(to)

	points := math.Min(len(match[2])-1, maxPoints)
	if match[2][0] == '-' {
		points *= -1
	}
	reason := match[4]

	record := &database.Points{
		From:   from,
		To:     to,
		Points: points,
		Reason: reason,
	}

	err = database.InsertPoints(record)
	if handleError(err, ev.Channel) {
		return
	}

	userPoints, err := database.GetPoints(to)
	if handleError(err, ev.Channel) {
		return
	}

	text := fmt.Sprintf("%s == %d (", to, userPoints)

	if points > 0 {
		text += "+"
	}
	text += strconv.Itoa(points)

	if reason != "" {
		text += fmt.Sprintf(" for %s", reason)
	}
	text += ")"

	b.SendMessage(text, ev.Channel)
}

func printLeaderboard(ev *slack.MessageEvent) {
	match := regexps.Leaderboard.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		return
	}

	limit := leaderboardLimit
	if match[1] != "" {
		var err error
		limit, err = strconv.Atoi(match[1])
		if handleError(err, ev.Channel) {
			return
		}
	}

	text := fmt.Sprintf("*top %d leaderboard*\n", limit)

	if hasWebUI {
		token, err := webui.GetToken()
		if err == nil {
			text += fmt.Sprintf("%sleaderboard/%d?token=%s\n", webUIURL, limit, token)
		}
	}

	leaderboard, err := database.GetLeaderboard(limit)
	if handleError(err, ev.Channel) {
		return
	}

	for i, user := range leaderboard {
		text += fmt.Sprintf("%d. %s == %d\n", i+1, munge(user.User), user.Points)
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(text, ev.Channel))
}

func parseUser(user string) (string, error) {
	if match := regexps.SlackUser.FindStringSubmatch(user); len(match) > 0 {
		return getUserNameByID(match[1])
	}

	return user, nil
}

func getUserNameByID(id string) (string, error) {
	userInfo, err := bot.GetUserInfo(id)
	if err != nil {
		return "", err
	}

	return userInfo.Name, nil
}
