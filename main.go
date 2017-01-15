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
		Motivate, GiveKarma, QueryKarma, Leaderboard, URL, SlackUser *regexp.Regexp
	}{
		Motivate:    karmaReg.GetMotivate(),
		GiveKarma:   karmaReg.GetGive(),
		QueryKarma:  karmaReg.GetQuery(),
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

// UserAliases is a map of alias -> main username
type UserAliases map[string]string

// Config contains all the necessary configs for karmabot.
type Config struct {
	Slack                       *Slack
	Debug, Motivate, Reactji    bool
	MaxPoints, LeaderboardLimit int
	Log                         *log.Log
	UI                          ui.Provider
	DB                          *database.DB
	UserBlacklist               StringList
	Aliases                     UserAliases
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
			case *slack.ReactionAddedEvent:
				go b.handleReactionAddedEvent(msg.Data.(*slack.ReactionAddedEvent))
			case *slack.ReactionRemovedEvent:
				go b.handleReactionRemovedEvent(msg.Data.(*slack.ReactionRemovedEvent))
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

// DMUser sends a message directly to a Slack user.
func (b *Bot) DMUser(message, user string) {
	_, _, channel, err := b.Config.Slack.Bot.OpenIMChannel(user)
	if err != nil {
		b.Config.Log.Err(err).KV("user", user).Error("could not open IM channel with user")
		return
	}

	b.Config.Slack.RTM.SendMessage(b.Config.Slack.RTM.NewOutgoingMessage(message, channel))
}

func (b *Bot) handleError(err error, channel string) bool {
	if err == nil {
		return false
	}

	fmt.Printf("%#v\n%T\n", err, err)
	b.Config.Log.Err(err).Error("error")
	if channel != "" {
		var message string
		if b.Config.Debug {
			message = err.Error()
		} else {
			message = "an error has occurred."
		}

		b.SendMessage(message, channel)
	}

	return true
}

func (b *Bot) handleReactionAddedEvent(ev *slack.ReactionAddedEvent) {
	if !b.Config.Reactji {
		return
	}

	var (
		points int
		reason string
	)
	switch ev.Reaction {
	case "+1":
		points = +1
		reason = "adding a :thumbsup: reactji"
	case "-1":
		points = -1
		reason = "adding a :thumbsdown: reactji"
	default:
		return
	}

	b.handleReactionEvent(ev.User, ev.ItemUser, reason, points)
}

func (b *Bot) handleReactionRemovedEvent(ev *slack.ReactionRemovedEvent) {
	if !b.Config.Reactji {
		return
	}

	var (
		points int
		reason string
	)
	switch ev.Reaction {
	case "+1":
		points = -1
		reason = "removing a :thumbsup: reactji"
	case "-1":
		points = 1
		reason = "removing a :thumbsdown: reactji"
	default:
		return
	}

	b.handleReactionEvent(ev.User, ev.ItemUser, reason, points)
}

func (b *Bot) handleReactionEvent(fromID, toID, reason string, points int) {
	from, err := b.getUserNameByID(fromID)
	if b.handleError(err, "") {
		return
	}
	to, err := b.getUserNameByID(toID)
	if b.handleError(err, "") {
		return
	}
	from, to = strings.ToLower(from), strings.ToLower(to)

	record := &database.Points{
		From:   from,
		To:     to,
		Points: points,
		Reason: reason,
	}

	err = b.Config.DB.InsertPoints(record)
	if b.handleError(err, "") {
		return
	}

	pointsMsg, err := b.getUserPointsMessage(to, reason, points)
	if b.handleError(err, "") {
		return
	}

	b.DMUser(pointsMsg, fromID)
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

	case regexps.QueryKarma.MatchString(ev.Text):
		b.queryKarma(ev)
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

	if _, blacklisted := b.Config.UserBlacklist[to]; blacklisted {
		b.Config.Log.KV("user", to).Info("user is blacklisted, ignoring karma command")
		return
	}

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

	pointsMsg, err := b.getUserPointsMessage(to, reason, points)
	if b.handleError(err, ev.Channel) {
		return
	}

	b.SendMessage(pointsMsg, ev.Channel)
}

func (b *Bot) getUserPointsMessage(name, reason string, points int) (string, error) {
	user, err := b.Config.DB.GetUser(name)
	if err != nil {
		return "", err
	}

	text := fmt.Sprintf("%s == %d (", name, user.Points)

	if points > 0 {
		text += "+"
	}
	text = fmt.Sprintf("%s%d", text, points)

	if reason != "" {
		text += fmt.Sprintf(" for %s", reason)
	}
	text += ")"

	return text, nil
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
		var err error
		user, err = b.getUserNameByID(match[1])
		if err != nil {
			return "", err
		}
	}

	// check if it is aliased
	if alias, ok := b.Config.Aliases[user]; ok {
		user = alias
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

func (b *Bot) queryKarma(ev *slack.MessageEvent) {
	match := regexps.QueryKarma.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		return
	}

	name, err := b.parseUser(match[1])
	if b.handleError(err, ev.Channel) {
		return
	}
	name = strings.ToLower(name)

	user, err := b.Config.DB.GetUser(name)
	switch {
	case err == database.ErrNoSuchUser:
		// override debug mode
		b.SendMessage(err.Error(), ev.Channel)
	case b.handleError(err, ev.Channel):
	default:
		b.SendMessage(fmt.Sprintf("%s == %d", user.Name, user.Points), ev.Channel)
	}
}
