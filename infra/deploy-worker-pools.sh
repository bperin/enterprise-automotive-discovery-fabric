#!/usr/bin/env sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
. "$root/config.env"

: "${PROJECT_ID:?set PROJECT_ID}"
: "${REGION:?set REGION}"
: "${IMAGE:?set IMAGE to an immutable image digest}"
: "${SERVICE_ACCOUNT:?set SERVICE_ACCOUNT}"
: "${RUNTIME_ENV:?set RUNTIME_ENV}"
: "${WORKER_POOL_SPECS:?set WORKER_POOL_SPECS as comma-separated name:command:instances entries}"

case "$IMAGE" in *@sha256:*) ;; *) echo "IMAGE must be immutable digest" >&2; exit 2;; esac

case "$SERVICE_ACCOUNT" in
  *@"$PROJECT_ID".iam.gserviceaccount.com)
    service_account_id=${SERVICE_ACCOUNT%@*}
    ;;
  *)
    echo "SERVICE_ACCOUNT must be an account in PROJECT_ID" >&2
    exit 2
    ;;
esac

# Runtime identity is created once and reused. This intentionally grants no
# project or data roles; operators manage least-privilege grants separately.
if ! gcloud iam service-accounts describe "$SERVICE_ACCOUNT" --project="$PROJECT_ID" >/dev/null 2>&1; then
  gcloud iam service-accounts create "$service_account_id" --project="$PROJECT_ID" \
    --display-name="$service_account_id Cloud Run worker"
fi

old_ifs=$IFS
IFS=,
for spec in $WORKER_POOL_SPECS; do
  field_ifs=$IFS
  IFS=:
  set -- $spec
  IFS=$field_ifs
  test "$#" = 3 || { echo "invalid worker pool spec: $spec" >&2; exit 2; }
  gcloud run worker-pools deploy "$1" --project="$PROJECT_ID" --region="$REGION" \
    --image="$IMAGE" --args="$2" --instances="$3" --service-account="$SERVICE_ACCOUNT" --set-env-vars="RUNTIME_ENV=$RUNTIME_ENV"
done
IFS=$old_ifs
