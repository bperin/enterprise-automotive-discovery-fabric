# Cloud Run worker pools

This generated project deploys background workers, not an HTTP service. Build
one immutable image and set `WORKER_POOL_SPECS` with explicit
`name:command:instances` entries before invoking `infra/deploy-worker-pools.sh`.

The deployment script idempotently creates the configured dedicated runtime
service account if it does not exist, then assigns it to every worker pool.
Workers use Application Default Credentials from that attached identity. The
script does not create service-account keys and does not grant any project,
dataset, topic, or subscription roles. Required least-privilege IAM grants are
operator-managed before deployment.
