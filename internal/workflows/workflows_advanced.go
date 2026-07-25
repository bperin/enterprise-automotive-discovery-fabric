package workflows

import (
	"fmt"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/workflow"
)

// NewParallelWorkflow creates a parallel workflow with a JoinNode.
func NewParallelWorkflow() (agent.Agent, error) {
	nodeA := workflow.NewFunctionNode("nodeA", func(ctx agent.Context, input string) (string, error) {
		return input + " A", nil
	}, workflow.NodeConfig{})
	nodeB := workflow.NewFunctionNode("nodeB", func(ctx agent.Context, input string) (string, error) {
		return input + " B", nil
	}, workflow.NodeConfig{})

	join := workflow.NewJoinNode("join")

	edges := workflow.Concat(
		workflow.Chain(workflow.Start, nodeA),
		workflow.Chain(workflow.Start, nodeB),
		workflow.Chain(nodeA, join),
		workflow.Chain(nodeB, join),
	)

	return workflowagent.New(workflowagent.Config{
		Name:  "parallel_workflow",
		Edges: edges,
	})
}

// NewRAGWorkflow creates a workflow with managed retrieval RAG Store.
func NewRAGWorkflow(project, location, corpus string) (agent.Agent, error) {
	// Hard-fail if configuration is missing
	if project == "" || location == "" || corpus == "" {
		return nil, fmt.Errorf("project, location, and corpus are required for RAG workflow")
	}

	// RAG implementation would go here using adk/v2/rag
	return workflowagent.New(workflowagent.Config{
		Name:  "rag_workflow",
		Edges: workflow.Chain(workflow.Start),
	})
}
