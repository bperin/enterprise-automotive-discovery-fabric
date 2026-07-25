# Deployment Guide

To deploy the generated service to Cloud Run:

1. Ensure you have the Google Cloud SDK installed and authenticated.
2. Build the image using Cloud Build:
   `gcloud builds submit --config infra/cloudbuild.yaml .`
3. Deploy the service:
   `./infra/gcloud.sh`

The deployment uses the configuration in `infra/cloudrun.yaml`.
The service uses a runtime service account and Secret Manager for configuration.
The service exposes a `/healthz` endpoint for readiness checks.
The image is tagged with the Git SHA for immutability.
