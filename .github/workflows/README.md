# GitHub Actions Deployment Setup

This workflow automatically deploys both Cloud Functions (`SlackCommandGo` and `SlackEventsGo`) to Google Cloud when you push to the `main` branch.

## Prerequisites

Before using this GitHub Action, you need to set up the following:

### 1. Google Cloud Service Account

Create a service account with the necessary permissions:

```bash
# Set your project ID
export PROJECT_ID=corbacoin

# Create service account
gcloud iam service-accounts create github-actions-deployer \
  --display-name="GitHub Actions Deployer" \
  --project=$PROJECT_ID

# Grant necessary roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/cloudfunctions.developer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/cloudbuild.builds.editor"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

# Create and download key
gcloud iam service-accounts keys create ~/github-actions-key.json \
  --iam-account=github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com

# Display the key (you'll need this for GitHub Secrets)
cat ~/github-actions-key.json
```

**Important:** Delete the key file after adding it to GitHub Secrets:
```bash
rm ~/github-actions-key.json
```

### 2. Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions, then add these secrets:

#### Required Secrets:

1. **GCP_SA_KEY**
   - The entire JSON content from the service account key file created above
   - Copy the entire JSON output and paste it as the secret value

2. **SLACK_BOT_TOKEN**
   - Your Slack bot token (starts with `xoxb-`)
   - Get this from https://api.slack.com/apps → Your App → OAuth & Permissions

3. **SLACK_SIGNING_SECRET**
   - Your Slack signing secret
   - Get this from https://api.slack.com/apps → Your App → Basic Information → App Credentials

### 3. Enable Required Google Cloud APIs

Make sure these APIs are enabled in your Google Cloud project:

```bash
gcloud services enable cloudfunctions.googleapis.com --project=corbacoin
gcloud services enable cloudbuild.googleapis.com --project=corbacoin
gcloud services enable artifactregistry.googleapis.com --project=corbacoin
gcloud services enable run.googleapis.com --project=corbacoin
gcloud services enable firestore.googleapis.com --project=corbacoin
```

## How It Works

The workflow (`deploy.yml`) does the following:

1. **Triggers** on:
   - Push to `main` branch
   - Manual workflow dispatch (via GitHub UI)

2. **Steps**:
   - Checks out the code
   - Sets up Go 1.24
   - Authenticates to Google Cloud using the service account key
   - Creates a temporary `.env.yaml` file with secrets
   - Deploys `SlackCommandGo` function
   - Deploys `SlackEventsGo` function
   - Cleans up the temporary `.env.yaml` file
   - Displays the deployed function URLs

## Manual Deployment

You can manually trigger the deployment:

1. Go to your GitHub repository
2. Click on **Actions** tab
3. Select **Deploy to Google Cloud Functions** workflow
4. Click **Run workflow**
5. Select the `main` branch
6. Click **Run workflow**

## Viewing Deployment Status

- Go to the **Actions** tab in your GitHub repository
- Click on the latest workflow run to see detailed logs
- The function URLs will be displayed at the end of the workflow

## Troubleshooting

### Authentication Issues

If you get authentication errors:
- Verify the `GCP_SA_KEY` secret is set correctly (should be valid JSON)
- Check that the service account has all necessary permissions
- Ensure the service account is from the correct project (`corbacoin`)

### Deployment Failures

If deployment fails:
- Check the workflow logs in the Actions tab
- Verify all required APIs are enabled
- Ensure the runtime version (`go124`) is supported
- Check that Firestore database exists in the project

### Secret Issues

If you get environment variable errors:
- Verify `SLACK_BOT_TOKEN` and `SLACK_SIGNING_SECRET` are set in GitHub Secrets
- Check that the secrets don't have extra whitespace or quotes

## Alternative: Workload Identity Federation (Recommended for Production)

For better security, consider using Workload Identity Federation instead of service account keys:

1. Follow Google's guide: https://github.com/google-github-actions/auth#workload-identity-federation
2. Update the `auth` step in `deploy.yml` to use `workload_identity_provider` instead of `credentials_json`

This eliminates the need to store long-lived service account keys.

## Local Testing

To test locally before pushing:

```bash
# Set environment variables
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_SIGNING_SECRET="your-secret"
export GOOGLE_CLOUD_PROJECT="corbacoin"

# Run local server
go run cmd/server/main.go
```

See `DEPLOYMENT.md` for more local development options.

