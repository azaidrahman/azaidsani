# Zaid's Blog

Personal website built with Go, deployed on Cloud Run.

## Local Development

```bash
go run main.go
```

Visit http://localhost:8090

## Docker

```bash
docker build -t blog .
docker run -p 8080:8080 -e PORT=8080 blog
```

Visit http://localhost:8080

## Deploy to GCP

### First-time setup

```bash
export PROJECT_ID=your-project-id
gcloud config set project $PROJECT_ID

gcloud services enable \
  cloudbuild.googleapis.com \
  run.googleapis.com \
  containerregistry.googleapis.com
```

### Manual deploy

```bash
gcloud builds submit --config cloudbuild.yaml
```

## Cost

With min-instances=0: **$0-2/month** unless you go viral.
