package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nlopes/slack"

	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/webui"
)

var (
	regexps = map[string]*regexp.Regexp{
		"motivate":    regexp.MustCompile(`^(?:\?|!)m\s+@?([^\s]+?)(?:\:\s)?$`),
		"karma":       regexp.MustCompile(`^@?([^\s]+?)(?:\:\s)?([\+]{2,}|[\-]{2,})((?: for)? (.*))?$`),
		"leaderboard": regexp.MustCompile(`^karma(?:bot)? (?:leaderboard|top|highscores) ?([0-9]+)?$`),
		"url":         regexp.MustCompile(`^karma(?:bot)? (?:url|web|link)?$`),
		"slackUser":   regexp.MustCompile(`^<@([A-Za-z0-9]+)>$`),
	}

	debug                       bool
	hasWebUI                    bool
	webUIURL                    string
	maxPoints, leaderboardLimit int
	bot                         *slack.Client
	rtm                         *slack.RTM
	ll                          *log.Logger
)

func main() {
	ll = log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags)

	var (
		flagToken            = flag.String("token", "", "slack RTM token")
		flagDBPath           = flag.String("db", "./db.sqlite3", "path to sqlite database")
		flagMaxPoints        = flag.Int("maxpoints", 6, "the maximum amount of points that users can give/take at once")
		flagLeaderboardLimit = flag.Int("leaderboardlimit", 10, "the default amount of users to list in the leaderboard")
		flagDebug            = flag.Bool("debug", false, "set debug mode")
		flagTOTP             = flag.String("totp", "", "totp key")
		flagWebUIPath        = flag.String("webuipath", "", "path to web UI files")
		flagListenAddr       = flag.String("listenaddr", "", "address to listen and serve the web ui on")
		flagWebUIURL         = flag.String("webuiurl", "", "url address for accessing the web ui")
	)

	flag.Parse()
	maxPoints = *flagMaxPoints
	leaderboardLimit = *flagLeaderboardLimit
	debug = *flagDebug
	hasWebUI = *flagWebUIPath != "" && *flagListenAddr != ""
	if hasWebUI {
		if *flagWebUIURL != "" {
			webUIURL = *flagWebUIURL
		} else {
			webUIURL = fmt.Sprintf("http://%s/", *flagListenAddr)
		}
	}

	if *flagToken == "" {
		ll.Fatalln("please pass the slack RTM token using the -token option")
	}

	database.Init(ll, *flagDBPath)

	bot = slack.New(*flagToken)
	slack.SetLogger(ll)
	bot.SetDebug(debug)

	rtm = bot.NewRTM()
	go rtm.ManageConnection()

	if hasWebUI {
		webui.Init(ll, *flagTOTP, *flagListenAddr, *flagWebUIPath)
		go webui.Listen()
	}

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				go handleMessage(msg)
			case *slack.ConnectedEvent:
				ll.Println("connected!")
				if debug {
					ll.Println("infos:", ev.Info)
					ll.Println("connection counter:", ev.ConnectionCount)
				}
			case *slack.RTMError:
				ll.Printf("Slack RTM error: %s\n", ev.Error())
			case *slack.InvalidAuthEvent:
				ll.Fatalln("invalid slack token")
			default:
			}

		}
	}
}

func handleMessage(msg slack.RTMEvent) {
	ev := msg.Data.(*slack.MessageEvent)

	if ev.Type != "message" {
		return
	}

	// convert motivates into karmabot syntax
	if match := regexps["motivate"].FindStringSubmatch(ev.Text); len(match) > 0 {
		ev.Text = match[1] + "++ for doing good work"
	}

	switch {
	case regexps["url"].MatchString(ev.Text):
		printURL(ev)

	case regexps["karma"].MatchString(ev.Text):
		givePoints(ev)

	case regexps["leaderboard"].MatchString(ev.Text):
		printLeaderboard(ev)
	}
}

func printURL(ev *slack.MessageEvent) {
	if !hasWebUI {
		rtm.SendMessage(rtm.NewOutgoingMessage("webui not enabled. please pass the `-webuipath`, `-webuiurl`, and `-listenaddr` options in order to enable the web ui", ev.Channel))
		return
	}

	token, err := webui.GetToken()
	if handleError(err, ev.Channel) {
		return
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("%s?token=%s", webUIURL, token), ev.Channel))
}

func givePoints(ev *slack.MessageEvent) {
	match := regexps["karma"].FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		return
	}

	from, err := getUserNameByID(ev.User)
	if handleError(err, ev.Channel) {
		return
	}
	to, err := parseUser(match[1])
	if handleError(err, ev.Channel) {
		return
	}
	to = strings.ToLower(to)

	points := min(len(match[2])-1, maxPoints)
	if match[2][0] == '-' {
		points *= -1
	}
	reason := match[4]

	record := database.Points{
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

	rtm.SendMessage(rtm.NewOutgoingMessage(text, ev.Channel))
}

func printLeaderboard(ev *slack.MessageEvent) {
	match := regexps["leaderboard"].FindStringSubmatch(ev.Text)
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
			text += fmt.Sprintf("%sleaderboard?token=%s&limit=%d\n", webUIURL, token, limit)
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

func handleError(err error, to string) bool {
	if err != nil {
		ll.Printf("Slack RTM error: %s\n", err.Error())
		rtm.SendMessage(rtm.NewOutgoingMessage("error: "+err.Error(), to))
		return true
	}

	return false
}

func parseUser(user string) (string, error) {
	if match := regexps["slackUser"].FindStringSubmatch(user); len(match) > 0 {
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
