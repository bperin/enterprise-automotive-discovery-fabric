#!/bin/bash
set -e
gcloud run deploy enterprise-search \
  --image us-docker.pkg.dev///enterprise-search: \
  --region  \
  --service-account  \
  --platform managed
