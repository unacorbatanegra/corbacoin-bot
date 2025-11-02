# Deployment Guide

## Code Structure

The project is structured to support both Cloud Functions deployment and local development:

- `function.go` (package `corbacoin`) - Cloud Functions entry point with `init()` function
- `cmd/server/main.go` (package `main`) - Local development server
- `main.go` (package `main`) - Deprecated, kept for reference (excluded from deployment)

**How it works:**
- Cloud Functions deploys `function.go` which registers HTTP handlers via `init()`
- Local development uses `cmd/server/main.go` which imports the corbacoin package
- The `.gcloudignore` file excludes `cmd/` and `main.go` from deployment

## Prerequisites

Before deploying, make sure these Google Cloud APIs are enabled:

```bash
# Enable required APIs
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable firestore.googleapis.com
```

### Firestore Database Setup

The bot uses Firestore to store user data. You need to create a Firestore database:

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project (`corbacoin`)
3. Navigate to **Firestore Database**
4. Click **Create database**
5. Choose **Production mode** or **Test mode** (for development)
6. Select a location (e.g., `us-central`)

**Note:** The bot will attempt to use the default (primary) Firestore database. If you need to use a named database, update the `FirestoreDatabase` variable in `config/config.go`.

## Create Artifact Registry Repository

Cloud Functions needs an Artifact Registry repository. Create one:

```bash
# Set your project
gcloud config set project corbacoin

# Create artifact registry repository for Cloud Functions
gcloud artifacts repositories create gcf-artifacts \
  --repository-format=docker \
  --location=us-central1 \
  --description="Docker repository for Cloud Functions"
```

Or use an existing repository if you already have one.

## Deploy Functions

### Recommended: Deploy with Gen2

```bash
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

### Option 2: Use environment variables from file

Create a `.env.yaml` file (don't commit this!):

```yaml
SLACK_BOT_TOKEN: "xoxb-your-token-here"
SLACK_SIGNING_SECRET: "your-secret-here"
```

Then deploy:

```bash
gcloud functions deploy SlackCommandGo \
  --gen2 \
  --runtime=go123 \
  --region=us-central1 \
  --source=. \
  --entry-point=SlackCommandGo \
  --trigger-http \
  --allow-unauthenticated \
  --project=corbacoin \
  --env-vars-file=.env.yaml
```

## Troubleshooting

### If you get the 'dockerRepository' error:

1. **Update gcloud SDK:**
   ```bash
   gcloud components update
   ```

2. **Check your region supports Gen2 functions:**
   ```bash
   gcloud functions regions list
   ```

3. **Try specifying the docker repository explicitly:**
   ```bash
   gcloud functions deploy SlackCommandGo \
     --gen2 \
     --runtime=go123 \
     --region=us-central1 \
     --source=. \
     --entry-point=SlackCommandGo \
     --trigger-http \
     --allow-unauthenticated \
     --project=corbacoin \
     --docker-registry=artifact-registry \
     --set-env-vars SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}",SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}"
   ```

4. **If still failing, try Gen1 (not recommended for new projects):**
   ```bash
   gcloud functions deploy SlackCommandGo \
     --runtime=go123 \
     --region=us-central1 \
     --source=. \
     --entry-point=SlackCommandGo \
     --trigger-http \
     --allow-unauthenticated \
     --project=corbacoin \
     --set-env-vars SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}",SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}"
   ```

### Check deployment status:

```bash
# List functions
gcloud functions list --gen2 --project=corbacoin

# Get function details
gcloud functions describe SlackCommandGo --gen2 --region=us-central1 --project=corbacoin

# View logs
gcloud functions logs read SlackCommandGo --gen2 --region=us-central1 --project=corbacoin
```

## After Deployment

Once deployed, you'll get URLs like:
- `https://us-central1-corbacoin.cloudfunctions.net/SlackCommandGo`
- `https://us-central1-corbacoin.cloudfunctions.net/SlackEventsGo`

Update your Slack app configuration with these URLs:
- **Slash Commands** → Request URL: `https://us-central1-corbacoin.cloudfunctions.net/SlackCommandGo`
- **Event Subscriptions** → Request URL: `https://us-central1-corbacoin.cloudfunctions.net/SlackEventsGo`

## Local Development

For local testing, use the `main.go` file which includes a `main()` function:

```bash
# Set environment variables
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_SIGNING_SECRET="your-secret"
export GOOGLE_CLOUD_PROJECT="corbacoin"

# Run locally
go run main.go
```

The local server will start on port 8080 and you can test the endpoints:
- `http://localhost:8080/SlackCommandGo`
- `http://localhost:8080/SlackEventsGo`

## Alternative: Deploy to Cloud Run

If Cloud Functions continues to give issues, you can deploy directly to Cloud Run:

```bash
# Build container
gcloud builds submit --tag gcr.io/corbacoin/corbacoin-bot --project=corbacoin

# Deploy to Cloud Run
gcloud run deploy corbacoin-bot \
  --image gcr.io/corbacoin/corbacoin-bot \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --project=corbacoin \
  --set-env-vars SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}",SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}"
```

For Cloud Run, you'd need to create a `Dockerfile` (let me know if you need this).

