package config

const (
	// InitialCoins is the number of coins a new user starts with
	InitialCoins = 5

	// LeaderboardLimit is the default number of users shown in the leaderboard
	LeaderboardLimit = 10

	// RequestTimestampTolerance is the maximum age of a request in seconds (5 minutes)
	RequestTimestampTolerance = 300
)

var (
	// SlackBotToken is the token for sending messages as the bot
	SlackBotToken string

	// SlackSigningSecret is used to verify Slack requests
	SlackSigningSecret string

	// FirestoreDatabase is the name of the Firestore database
	FirestoreDatabase = "corbacoin-database"
)

