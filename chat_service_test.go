package karmabot

import (
	"github.com/nlopes/slack"
)

type TestChatService struct {
	IncomingEvents chan slack.RTMEvent

	SentMessages []*slack.OutgoingMessage
	id           int
}

func newTestChatService() ChatService {
	return &TestChatService{}
}

func (t *TestChatService) IncomingEventsChan() chan slack.RTMEvent {
	return t.IncomingEvents
}

func (t *TestChatService) OpenIMChannel(user string) (bool, bool, string, error) {
	return true, true, user, nil
}

func (t *TestChatService) GetUserInfo(user string) (*slack.User, error) {
	return &slack.User{
		ID:   user,
		Name: user,
	}, nil
}

func (t *TestChatService) NewOutgoingMessage(text string, channel string, options ...slack.RTMsgOption) *slack.OutgoingMessage {
	t.id++
	return &slack.OutgoingMessage{
		ID:      t.id,
		Type:    "message",
		Channel: channel,
		Text:    text,
	}
}

func (t *TestChatService) SendMessage(m *slack.OutgoingMessage) {
	t.SentMessages = append(t.SentMessages, m)
}

func (t *TestChatService) PostEphemeral(channelID, userID string, options ...slack.MsgOption) (string, error) {
	// run options
	_, values, _ := slack.UnsafeApplyMsgOptions("", "", options...)
	message := values.Get("text")

	t.SendMessage(
		t.NewOutgoingMessage(
			message, "user",
		),
	)

	return "", nil
}
