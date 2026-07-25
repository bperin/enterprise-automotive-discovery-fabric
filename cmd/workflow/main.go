package main

import (
	"context"
	"flag"
	"log"
	"os"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"

	"enterprise-search/internal/workflows"
)

func main() {
	name := flag.String("name", "sequential", "workflow name: sequential|routing|loop|hitl|parallel|rag")
	project := flag.String("project", "", "GCP project ID")
	location := flag.String("location", "", "GCP location")
	corpus := flag.String("corpus", "", "RAG corpus ID")
	flag.Parse()

	ctx := context.Background()

	var wf agent.Agent
	var err error

	switch *name {
	case "sequential":
		wf, err = workflows.NewSequentialWorkflow()
	case "routing":
		wf, err = workflows.NewRoutingWorkflow()
	case "loop":
		wf, err = workflows.NewLoopWorkflow()
	case "hitl":
		wf, err = workflows.NewHITLWorkflow()
	case "parallel":
		wf, err = workflows.NewParallelWorkflow()
	case "rag":
		wf, err = workflows.NewRAGWorkflow(*project, *location, *corpus)
	default:
		log.Fatalf("unknown workflow: %s", *name)
	}

	if err != nil {
		log.Fatalf("failed to create workflow: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(wf),
	}
	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
