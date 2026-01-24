# My Blog

Dead-simple blog using plain HTML + Cloud Run.

## Local Development

```bash
# Option 1: Python's built-in server
cd public && python3 -m http.server 8080

# Option 2: Docker (matches production)
docker build -t blog .
docker run -p 8080:8080 blog
```

Then visit http://localhost:8080

## Deploy to GCP

### First-time setup

```bash
# Set your project
export PROJECT_ID=your-project-id
gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable \
  cloudbuild.googleapis.com \
  run.googleapis.com \
  containerregistry.googleapis.com

# Grant Cloud Build permission to deploy to Cloud Run
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com" \
  --role="roles/run.admin"
gcloud iam service-accounts add-iam-policy-binding \
  ${PROJECT_NUMBER}-compute@developer.gserviceaccount.com \
  --member="serviceAccount:${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Connect your GitHub repo to Cloud Build
# Go to: https://console.cloud.google.com/cloud-build/triggers
# Click "Connect Repository" and follow the steps
# Create a trigger on push to main branch
```

### Manual deploy (if you want to test before setting up CI/CD)

```bash
gcloud builds submit --config cloudbuild.yaml
```

## Writing a new post

1. Create a new file: `public/posts/YYYY-MM-DD-slug.html`
2. Copy the structure from an existing post
3. Add it to `public/posts/index.html` and `public/index.html`
4. `git push` — Cloud Build handles the rest

## Custom domain (optional)

```bash
# Map your domain to Cloud Run
gcloud run domain-mappings create \
  --service=blog \
  --domain=yourdomain.com \
  --region=us-central1

# Then add the DNS records it tells you to add
```

## Cost

With min-instances=0, you only pay when someone visits:
- Cloud Run: Free tier covers 2M requests/month
- Container Registry: ~$0.026/GB storage
- Cloud Build: Free tier covers 120 build-minutes/day

Realistically: **$0-2/month** unless you go viral.
