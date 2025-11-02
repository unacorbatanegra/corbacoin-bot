package models

// User represents a user in the system with their coin balance
type User struct {
	Username string `firestore:"username"`
	Coins    int    `firestore:"coins"`
}

// SlackCommandRequest represents an incoming Slack slash command
type SlackCommandRequest struct {
	Command     string `json:"command"`
	Text        string `json:"text"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ResponseURL string `json:"response_url"`
	ChannelID   string `json:"channel_id"`
}

// SlackResponse represents a response to a Slack command
type SlackResponse struct {
	Text         string `json:"text"`
	ResponseType string `json:"response_type"`
}

// SlackMessage represents a message to post to Slack
type SlackMessage struct {
	Channel  string `json:"channel"`
	Text     string `json:"text"`
	ThreadTS string `json:"thread_ts,omitempty"`
}

// SlackEventPayload represents an incoming Slack event
type SlackEventPayload struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"`
	Event     SlackEventInner `json:"event,omitempty"`
}

// SlackEventInner represents the inner event data
type SlackEventInner struct {
	Type     string `json:"type"`
	Channel  string `json:"channel"`
	User     string `json:"user"`
	Text     string `json:"text"`
	TS       string `json:"ts"`
	ThreadTS string `json:"thread_ts,omitempty"`
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	Success bool
	Message string
}

