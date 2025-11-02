# Corbacoin Bot

Slack bot for Corbacoin currency system.

## Usage

**Slash Commands:**
- `/balance` - Check your balance
- `/send @user amount` - Send corbacoins to another user
- `/leaderboard` - View top 10 users

**Mentions:**
- `@CorbacoinBot balance` - Check your balance
- `@CorbacoinBot send @user amount` - Send corbacoins
- `@CorbacoinBot leaderboard` - View leaderboard
- `@CorbacoinBot help` - Show help

## Deploy
```bash
gcloud functions deploy SlackEventsGo \
  --gen2 \
  --runtime=go124 \
  --region=us-central1 \
  --source=. \
  --entry-point=SlackEventsGo \
  --trigger-http \
  --allow-unauthenticated \
  --project=corbacoin \
  --env-vars-file=.env.yaml

gcloud functions deploy SlackCommandGo \
  --gen2 \
  --runtime=go124 \
  --region=us-central1 \
  --source=. \
  --entry-point=SlackCommandGo \
  --trigger-http \
  --allow-unauthenticated \
  --project=corbacoin \
  --env-vars-file=.env.yaml
```