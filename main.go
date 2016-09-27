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
				handleMessage(msg)
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

	userInfo, err := bot.GetUserInfo(ev.User)
	if handleError(err, ev.Channel) {
		return
	}
	from := userInfo.Name
	to := strings.ToLower(match[1])
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

// apparently you should do this for ints
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

var mungeMap = map[byte]rune{
	'a': '\u00e4',
	'b': '\u0411',
	'c': '\u010b',
	'd': '\u0111',
	'e': '\u00eb',
	'f': '\u0192',
	'g': '\u0121',
	'h': '\u0127',
	'i': '\u00ed',
	'j': '\u0135',
	'k': '\u0137',
	'l': '\u013a',
	'm': '\u1e41',
	'n': '\u00f1',
	'o': '\u00f6',
	'p': '\u03c1',
	'q': '\u02a0',
	'r': '\u0157',
	's': '\u0161',
	't': '\u0163',
	'u': '\u00fc',
	'v': 'v',
	'w': '\u03c9',
	'x': '\u03c7',
	'y': '\u00ff',
	'z': '\u017a',
	'A': '\u00c5',
	'B': '\u0392',
	'C': '\u00c7',
	'D': '\u010e',
	'E': '\u0112',
	'F': '\u1e1e',
	'G': '\u0120',
	'H': '\u0126',
	'I': '\u00cd',
	'J': '\u0134',
	'K': '\u0136',
	'L': '\u0139',
	'M': '\u039c',
	'N': '\u039d',
	'O': '\u00d6',
	'P': '\u0420',
	'Q': '\uff31',
	'R': '\u0156',
	'S': '\u0160',
	'T': '\u0162',
	'U': '\u016e',
	'V': '\u1e7e',
	'W': '\u0174',
	'X': '\u03a7',
	'Y': '\u1ef2',
	'Z': '\u017b',
}

func munge(str string) string {
	if len(str) < 1 {
		return str
	}

	first := str[0]
	out := []rune(str)
	if munged, ok := mungeMap[first]; ok {
		out[0] = munged
	}

	return string(out)
}
