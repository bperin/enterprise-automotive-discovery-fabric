package secretmanager

import (
	"context"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"enterprise-search/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client defines the subset of Secret Manager API methods we need.
type Client interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

// clientWrapper wraps the real GCP Secret Manager client to implement our Client interface.
type clientWrapper struct {
	client *secretmanager.Client
}

func (w *clientWrapper) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return w.client.AccessSecretVersion(ctx, req)
}

func (w *clientWrapper) Close() error {
	return w.client.Close()
}

// Adapter implements config.SecretResolver using Google Cloud Secret Manager.
type Adapter struct {
	client Client
}

var _ config.SecretResolver = (*Adapter)(nil)

// NewAdapter creates a new Secret Manager adapter using the provided client.
func NewAdapter(client Client) *Adapter {
	return &Adapter{client: client}
}

// NewProductionAdapter creates a new Secret Manager adapter using the official GCP client with ADC.
func NewProductionAdapter(ctx context.Context) (*Adapter, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create secret manager client: %w", err)
	}
	return &Adapter{client: &clientWrapper{client: client}}, nil
}

// Resolve resolves a Secret Manager reference string to its secret value.
func (a *Adapter) Resolve(ctx context.Context, ref string) (string, error) {
	project, secret, version, err := config.ParseSecretRef(ref)
	if err != nil {
		return "", &config.MalformedReferenceError{Reference: ref, Err: err}
	}

	name := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, secret, version)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	resp, err := a.client.AccessSecretVersion(ctx, req)
	if err != nil {
		// Check if the error is a gRPC NOT_FOUND error
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return "", &config.SecretNotFoundError{Reference: ref, Err: err}
		}
		// Also check if the error message contains "not found" or "NotFound"
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "NotFound") {
			return "", &config.SecretNotFoundError{Reference: ref, Err: err}
		}
		return "", &config.ProviderError{Reference: ref, Err: err}
	}

	if resp.Payload == nil || resp.Payload.Data == nil {
		return "", &config.SecretNotFoundError{Reference: ref, Err: fmt.Errorf("empty payload in secret response")}
	}

	return string(resp.Payload.Data), nil
}
