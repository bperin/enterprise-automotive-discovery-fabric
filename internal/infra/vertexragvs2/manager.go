package vertexragvs2

import (
	aiplatform "cloud.google.com/go/aiplatform/apiv1beta1"
	pb "cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"strings"
)

const DefaultEmbeddingModel = "gemini-embedding-001"

type Config struct {
	Project        string
	Location       string
	EmbeddingModel string
}

func (c Config) validate() error {
	if strings.TrimSpace(c.Project) == "" || strings.TrimSpace(c.Location) == "" {
		return fmt.Errorf("RAG-managed Vector Search 2 requires project and location")
	}
	return nil
}
func (c Config) model() string {
	m := strings.TrimSpace(c.EmbeddingModel)
	if m == "" {
		m = DefaultEmbeddingModel
	}
	if strings.HasPrefix(m, "projects/") || strings.HasPrefix(m, "publishers/") {
		return m
	}
	return fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s", c.Project, c.Location, m)
}

type Corpus struct {
	Name           string
	CollectionName string
}

// Manager provisions a RAG corpus whose Vector Search 2 collection is created and owned by RAG Engine.
type Manager struct {
	client *aiplatform.VertexRagDataClient
	config Config
}

func NewManager(ctx context.Context, c Config) (*Manager, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	client, err := aiplatform.NewVertexRagDataClient(ctx, option.WithEndpoint(c.Location+"-aiplatform.googleapis.com:443"))
	if err != nil {
		return nil, fmt.Errorf("create RAG-managed Vector Search 2 client with ADC: %w", err)
	}
	return &Manager{client: client, config: c}, nil
}
func (m *Manager) Close() error { return m.client.Close() }
func (m *Manager) CreateCorpus(ctx context.Context, name string) (*Corpus, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("corpus display name is required")
	}
	req := &pb.CreateRagCorpusRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", m.config.Project, m.config.Location),
		RagCorpus: &pb.RagCorpus{
			DisplayName: name,
			BackendConfig: &pb.RagCorpus_VectorDbConfig{
				VectorDbConfig: &pb.RagVectorDbConfig{
					VectorDb: &pb.RagVectorDbConfig_RagManagedVertexVectorSearch_{
						RagManagedVertexVectorSearch: &pb.RagVectorDbConfig_RagManagedVertexVectorSearch{},
					},
					RagEmbeddingModelConfig: &pb.RagEmbeddingModelConfig{
						ModelConfig: &pb.RagEmbeddingModelConfig_VertexPredictionEndpoint_{
							VertexPredictionEndpoint: &pb.RagEmbeddingModelConfig_VertexPredictionEndpoint{Endpoint: m.config.model()},
						},
					},
				},
			},
		},
	}
	op, err := m.client.CreateRagCorpus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create RAG-managed Vector Search 2 corpus: %w", err)
	}
	corpus, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for RAG-managed Vector Search 2 corpus: %w", err)
	}
	return &Corpus{Name: corpus.GetName(), CollectionName: corpus.GetVectorDbConfig().GetRagManagedVertexVectorSearch().GetCollectionName()}, nil
}
func (m *Manager) ImportGCS(ctx context.Context, corpusName string, uris []string) error {
	if corpusName == "" || len(uris) == 0 {
		return fmt.Errorf("corpus name and GCS URIs are required")
	}
	for _, uri := range uris {
		if !strings.HasPrefix(uri, "gs://") {
			return fmt.Errorf("GCS URI %q must start with gs://", uri)
		}
	}
	op, err := m.client.ImportRagFiles(ctx, &pb.ImportRagFilesRequest{Parent: corpusName, ImportRagFilesConfig: &pb.ImportRagFilesConfig{ImportSource: &pb.ImportRagFilesConfig_GcsSource{GcsSource: &pb.GcsSource{Uris: uris}}}})
	if err != nil {
		return fmt.Errorf("import RAG files: %w", err)
	}
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for RAG file import: %w", err)
	}
	return nil
}
