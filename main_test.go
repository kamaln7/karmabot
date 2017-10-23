package karmabot

import (
	"testing"
	"time"

	"github.com/kamaln7/karmabot/database"
	"github.com/nlopes/slack"
)

func TestNew(t *testing.T) {
	cfg := &Config{}

	b := New(cfg)
	if b == nil {
		t.Fatalf("New(cfg) returned nil; wanted *Bot")
	}

	if b.Config != cfg {
		t.Errorf("New(cfg): returned Bot with incorrect config")
	}
}

func newBot(cfg *Config) (*Bot, *TestChatService, *TestDatabase) {
	cs := &TestChatService{
		IncomingEvents: make(chan slack.RTMEvent),
	}
	db := &TestDatabase{}
	db.InsertPoints(&database.Points{
		From:   "point_giver",
		To:     "onehundred_points",
		Points: 100,
		Reason: "for being a swell guy",
	})
	cfg.Slack = cs
	cfg.DB = db
	return New(cfg), cs, db
}

func TestListen(t *testing.T) {
	b, cs, _ := newBot(&Config{})
	hasExited := false
	hasStarted := make(chan int)
	go func() {
		close(hasStarted)
		b.Listen()
		hasExited = true
	}()
	<-hasStarted
	time.Sleep(1 * time.Millisecond)
	if hasExited {
		t.Errorf("Listen: exited immediately")
	}
	close(cs.IncomingEvents)
	time.Sleep(1 * time.Millisecond)
	if !hasExited {
		t.Errorf("Listen: did not exit after closing incoming events channel")
	}

	// TODO: To properly test Listen, it needs to be decoupled further from what it actually does.
}

func TestHandleSlackEvent(t *testing.T) {
	tt := []struct {
		Name                 string
		ReacjiDisabled       bool
		ReactionAddedEvent   *slack.ReactionAddedEvent
		ReactionRemovedEvent *slack.ReactionRemovedEvent
		MessageEvent         *slack.MessageEvent
		ExpectMessage        string
		ShouldHavePoints     int
	}{
		{
			Name:           "+1 added with reacji disabled",
			ReacjiDisabled: true,
			ReactionAddedEvent: &slack.ReactionAddedEvent{
				Type:     "reaction_added",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "+1",
			},
			ShouldHavePoints: 100,
		},
		{
			Name: "+1 added with reacji enabled",
			ReactionAddedEvent: &slack.ReactionAddedEvent{
				Type:     "reaction_added",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "+1",
			},
			ExpectMessage:    "onehundred_points == 101 (+1 for adding a :+1: reactji)",
			ShouldHavePoints: 101,
		},
		{
			Name: "-1 added with reacji enabled",
			ReactionAddedEvent: &slack.ReactionAddedEvent{
				Type:     "reaction_added",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "-1",
			},
			ExpectMessage:    "onehundred_points == 99 (-1 for adding a :-1: reactji)",
			ShouldHavePoints: 99,
		},
		{
			Name: "cat added with reacji enabled",
			ReactionAddedEvent: &slack.ReactionAddedEvent{
				Type:     "reaction_added",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "cat",
			},
			ShouldHavePoints: 100,
		},

		{
			Name:           "+1 removed with reacji disabled",
			ReacjiDisabled: true,
			ReactionRemovedEvent: &slack.ReactionRemovedEvent{
				Type:     "reaction_removed",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "+1",
			},
			ShouldHavePoints: 100,
		},
		{
			Name: "+1 removed with reacji enabled",
			ReactionRemovedEvent: &slack.ReactionRemovedEvent{
				Type:     "reaction_removed",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "+1",
			},
			ExpectMessage:    "onehundred_points == 99 (-1 for removing a :+1: reactji)",
			ShouldHavePoints: 99,
		},
		{
			Name: "-1 removed with reacji enabled",
			ReactionRemovedEvent: &slack.ReactionRemovedEvent{
				Type:     "reaction_removed",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "-1",
			},
			ExpectMessage:    "onehundred_points == 101 (+1 for removing a :-1: reactji)",
			ShouldHavePoints: 101,
		},
		{
			Name: "cat removed with reacji enabled",
			ReactionRemovedEvent: &slack.ReactionRemovedEvent{
				Type:     "reaction_removed",
				User:     "user",
				ItemUser: "onehundred_points",
				Reaction: "cat",
			},
			ShouldHavePoints: 100,
		},
		{
			Name: "should tell user about their sick karma events from the past",
			MessageEvent: &slack.MessageEvent{
				Msg: slack.Msg{
					Type:    "message",
					Text:    "karmabot throwback",
					Channel: "user",
					User:    "onehundred_points",
				},
			},
			ExpectMessage:    "önehundred_points received 100 points from ρoint_giver now for for being a swell guy",
			ShouldHavePoints: 100,
		},
	}

	for _, tc := range tt {
		upvote, downvote := make(StringList, 1), make(StringList, 1)
		upvote.Set("+1")
		downvote.Set("-1")

		b, cs, db := newBot(&Config{
			Reactji: &ReactjiConfig{
				Enabled:  !tc.ReacjiDisabled,
				Upvote:   upvote,
				Downvote: downvote,
			},
		})

		if tc.ReactionAddedEvent != nil {
			b.handleReactionAddedEvent(tc.ReactionAddedEvent)
		}
		if tc.ReactionRemovedEvent != nil {
			b.handleReactionRemovedEvent(tc.ReactionRemovedEvent)
		}
		if tc.MessageEvent != nil {
			b.handleMessageEvent(tc.MessageEvent)
		}

		if len(cs.SentMessages) != 0 && tc.ExpectMessage == "" {
			t.Errorf("%s: sent unexpected message %#v", tc.Name, cs.SentMessages[0])
		} else if len(cs.SentMessages) == 0 && tc.ExpectMessage != "" {
			t.Errorf("%s: did not send expected message %v", tc.Name, tc.ExpectMessage)
		} else if len(cs.SentMessages) > 1 {
			t.Errorf("%s: sent too many messages: %v", tc.Name, cs.SentMessages)
		} else if tc.ExpectMessage != "" {
			msg := cs.SentMessages[0]
			if msg.Text != tc.ExpectMessage {
				t.Errorf("%s: sent message %q; want %q", tc.Name, msg.Text, tc.ExpectMessage)
			}
			if msg.Channel != "user" {
				t.Errorf("%s: sent message to %q; want to send message to %q", tc.Name, msg.Channel, "user")
			}
		}

		u, err := db.GetUser("onehundred_points")
		if err != nil {
			t.Fatalf("%s: db.GetUser: %v", tc.Name, err)
		}

		if u.Points != tc.ShouldHavePoints {
			t.Errorf("%s: user %v has %v points; want %v", tc.Name, "onehundred_points", u.Points, tc.ShouldHavePoints)
		}
	}
}
