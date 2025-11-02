# Quick Deploy Guide

## What Was Fixed

### Previous Error
```
import "github.com/unacorbatanegra/corbacoin-bot" is a program, not an importable package
```

### Solution Applied
Changed from `package main` to `package corbacoin` in `function.go` to make it importable by Cloud Functions.

### Runtime Error Fixed
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**Root Cause:** Firestore client initialization was failing silently, leaving `database.Client` as `nil`.

**Fix Applied:**
- Added proper error checking with `log.Fatal()` if Firestore initialization fails
- Try default database first, fallback to named database
- Better logging to identify initialization issues

## Deploy Now

```bash
# Set your environment (if not already set)
export SLACK_BOT_TOKEN="xoxb-your-token-here"
export SLACK_SIGNING_SECRET="your-signing-secret-here"

# Deploy SlackCommandGo
gcloud functions deploy SlackCommandGo \
  --gen2 \
  --runtime=go123 \
  --region=us-central1 \
  --source=. \
  --entry-point=SlackCommandGo \
  --trigger-http \
  --allow-unauthenticated \
  --project=corbacoin \
  --set-env-vars SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}",SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}"

# Deploy SlackEventsGo
gcloud functions deploy SlackEventsGo \
  --gen2 \
  --runtime=go123 \
  --region=us-central1 \
  --source=. \
  --entry-point=SlackEventsGo \
  --trigger-http \
  --allow-unauthenticated \
  --project=corbacoin \
  --set-env-vars SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}",SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}"
```

## Verify Deployment

```bash
# Check function status
gcloud functions describe SlackCommandGo --gen2 --region=us-central1 --project=corbacoin

# View logs
gcloud functions logs read SlackCommandGo --gen2 --region=us-central1 --project=corbacoin --limit=50

# Test the function
curl -X POST https://us-central1-corbacoin.cloudfunctions.net/SlackCommandGo
```

## Important Notes

1. **Firestore Database**: Make sure you have created a Firestore database in the Firebase Console
2. **Environment Variables**: The function will receive `GOOGLE_CLOUD_PROJECT` automatically from Cloud Functions
3. **Logs**: Check logs if the function fails to start - look for Firestore initialization errors

## Local Development

```bash
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_SIGNING_SECRET="your-secret"
export GOOGLE_CLOUD_PROJECT="corbacoin"

# Run local server
go run cmd/server/main.go
```

Test locally at:
- http://localhost:8080/SlackCommandGo
- http://localhost:8080/SlackEventsGo

