package main

import (
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nlopes/slack"

	"github.com/kamaln7/karmabot/database"
)

var (
	motivateRegexp    = regexp.MustCompile(`^(?:\?|!)m\s+@?([^\s]+?)(?:\:\s)?$`)
	karmaRegexp       = regexp.MustCompile(`^@?([^\s]+?)(?:\:\s)?([\+]{2,}|[\-]{2,})((?: for)? (.*))?$`)
	leaderboardRegexp = regexp.MustCompile(`^karma(?:bot)? (?:leaderboard|top|highscores) ?([0-9]+)?$`)
	slackUserRegexp   = regexp.MustCompile(`^<@([A-Za-z0-9]+)>$`)

	maxPoints, leaderboardLimit int
	bot                         *slack.Client
	rtm                         *slack.RTM
)

func main() {
	ll := log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags)

	var (
		flagToken            = flag.String("token", "", "slack RTM token")
		flagDBPath           = flag.String("db", "./db.sqlite3", "path to sqlite database")
		flagMaxPoints        = flag.Int("maxpoints", 6, "the maximum amount of points that users can give/take at once")
		flagLeaderboardLimit = flag.Int("leaderboardLimit", 10, "the default amount of users to list in the leaderboard")
	)

	flag.Parse()
	maxPoints = *flagMaxPoints
	leaderboardLimit = *flagLeaderboardLimit
	if *flagToken == "" {
		ll.Fatalln("please pass the slack RTM token using the -token option")
	}

	database.Init(ll, *flagDBPath)

	bot = slack.New(*flagToken)
	slack.SetLogger(ll)
	bot.SetDebug(true)

	rtm = bot.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				go handleMessage(msg)
			case *slack.ConnectedEvent:
				ll.Println("connected!")
				ll.Println("infos:", ev.Info)
				ll.Println("connection counter:", ev.ConnectionCount)
			case *slack.RTMError:
				ll.Printf("RTM error: %s\n", ev.Error())
			case *slack.InvalidAuthEvent:
				ll.Fatalln("invalid slack token")
				break Loop
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
	if match := motivateRegexp.FindStringSubmatch(ev.Text); len(match) > 0 {
		ev.Text = match[1] + "++ for doing good work"
	}

	if karmaRegexp.MatchString(ev.Text) {
		givePoints(ev)
	}

	if leaderboardRegexp.MatchString(ev.Text) {
		printLeaderboard(ev)
	}
}

func givePoints(ev *slack.MessageEvent) {
	match := karmaRegexp.FindStringSubmatch(ev.Text)
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

	givenPointsText := ""
	if points > 0 {
		givenPointsText = "+"
	}
	givenPointsText += strconv.Itoa(points)

	text := to + " == " + strconv.Itoa(userPoints) + " (" + givenPointsText
	if reason != "" {
		text += " for " + reason
	}
	text += ")"

	rtm.SendMessage(rtm.NewOutgoingMessage(text, ev.Channel))
}

func printLeaderboard(ev *slack.MessageEvent) {
	match := leaderboardRegexp.FindStringSubmatch(ev.Text)
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
	text := "top " + strconv.Itoa(limit) + " leaderboard\n"

	leaderboard, err := database.GetLeaderboard(limit)
	if handleError(err, ev.Channel) {
		return
	}

	for _, user := range leaderboard {
		text += munge(user.User) + " == " + strconv.Itoa(user.Points) + "\n"
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(text, ev.Channel))
}

func handleError(err error, to string) bool {
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage("error: "+err.Error(), to))
		return true
	}

	return false
}

func parseUser(user string) (string, error) {
	if match := slackUserRegexp.FindStringSubmatch(user); len(match) > 0 {
		return getUserNameByID(match[1])
	} else {
		return user, nil
	}
}

func getUserNameByID(id string) (string, error) {
	userInfo, err := bot.GetUserInfo(id)
	if err != nil {
		return "", err
	}

	return userInfo.Name, nil
}
