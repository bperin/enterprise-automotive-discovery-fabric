package ragworkflow

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/geminitool"
	"google.golang.org/genai"
)

// VertexRAGStoreConfig selects a corpus for native Vertex RAG Store grounding.
type VertexRAGStoreConfig struct {
	Project    string
	Location   string
	CorpusName string
	Model      string
	TopK       int32
}

func (c VertexRAGStoreConfig) validate() error {
	if strings.TrimSpace(c.Project) == "" || strings.TrimSpace(c.Location) == "" || strings.TrimSpace(c.CorpusName) == "" || strings.TrimSpace(c.Model) == "" || c.TopK <= 0 {
		return fmt.Errorf("Vertex RAG Store requires project, location, corpus name, model, and positive topK")
	}
	return nil
}

// NewVertexRAGStoreAgent is the normal model-time RAG path. ADK supplies a
// native VertexRAGStore retrieval tool to Gemini; no custom retrieval HTTP call
// or manually assembled context prompt sits between the user and the model.
func NewVertexRAGStoreAgent(ctx context.Context, c VertexRAGStoreConfig) (agent.Agent, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	llm, err := gemini.NewModel(ctx, c.Model, &genai.ClientConfig{Backend: genai.BackendVertexAI, Project: c.Project, Location: c.Location})
	if err != nil {
		return nil, fmt.Errorf("create Vertex Gemini model with ADC: %w", err)
	}
	return newVertexRAGStoreAgent(llm, c)
}
func newVertexRAGStoreAgent(llm model.LLM, c VertexRAGStoreConfig) (agent.Agent, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	return llmagent.New(llmagent.Config{Name: "vertex_rag_store", Description: "answers from a Vertex RAG corpus with native grounding", Model: llm, Instruction: "Answer from the configured Vertex RAG Store. State when the corpus does not support an answer and preserve grounding citations.", Tools: []tool.Tool{NativeRAGTool(c)}})
}
func NativeRAGTool(c VertexRAGStoreConfig) tool.Tool {
	return geminitool.New("vertex_rag_store", "native Vertex RAG Store retrieval", &genai.Tool{Retrieval: &genai.Retrieval{VertexRAGStore: &genai.VertexRAGStore{RAGResources: []*genai.VertexRAGStoreRAGResource{{RAGCorpus: c.CorpusName}}, RAGRetrievalConfig: &genai.RAGRetrievalConfig{TopK: &c.TopK}}}})
}
