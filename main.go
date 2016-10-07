package karmabot

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
)

var (
	motivateRegexp    = regexp.MustCompile(`^(?:\?|!)m\s+@?([^\s]+?)(?:\:\s)?$`)
	karmaRegexp       = regexp.MustCompile(`^@?([^\s]+?)(?:\:\s)?([\+]{2,}|[\-]{2,})((?: for)? (.*))?$`)
	leaderboardRegexp = regexp.MustCompile(`^karma(?:bot)? (?:leaderboard|top|highscores) ?([0-9]+)?$`)
	slackUserRegexp   = regexp.MustCompile(`^<@([A-Za-z0-9]+)>$`)

	debug                       bool
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
	)

	flag.Parse()
	maxPoints = *flagMaxPoints
	leaderboardLimit = *flagLeaderboardLimit
	debug = *flagDebug
	if *flagToken == "" {
		ll.Fatalln("please pass the slack RTM token using the -token option")
	}

	database.Init(ll, *flagDBPath)

	bot = slack.New(*flagToken)
	slack.SetLogger(ll)
	bot.SetDebug(debug)

	rtm = bot.NewRTM()
	go rtm.ManageConnection()

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
	text := fmt.Sprintf("top %d leaderboard\n", limit)

	leaderboard, err := database.GetLeaderboard(limit)
	if handleError(err, ev.Channel) {
		return
	}

	for _, user := range leaderboard {
		text += fmt.Sprintf("%s == %d\n", munge(user.User), user.Points)
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
	if match := slackUserRegexp.FindStringSubmatch(user); len(match) > 0 {
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
