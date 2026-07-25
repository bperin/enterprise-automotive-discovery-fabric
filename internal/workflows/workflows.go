package workflows

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

// NewSequentialWorkflow creates a simple sequential workflow.
func NewSequentialWorkflow() (agent.Agent, error) {
	upperFn := func(ctx agent.Context, input string) (string, error) {
		return strings.ToUpper(input), nil
	}
	nodeA := workflow.NewFunctionNode("upper", upperFn, workflow.NodeConfig{})

	return workflowagent.New(workflowagent.Config{
		Name:  "sequential_workflow",
		Edges: workflow.Chain(workflow.Start, nodeA),
	})
}

// NewRoutingWorkflow creates a routing workflow.
func NewRoutingWorkflow() (agent.Agent, error) {
	classify := func(ctx agent.Context, input string, emit func(*session.Event) error) (any, error) {
		ev := session.NewEvent(ctx, ctx.InvocationID())
		if strings.Contains(input, "?") {
			ev.Routes = []string{"question"}
		} else {
			ev.Routes = []string{"statement"}
		}
		if err := emit(ev); err != nil {
			return nil, err
		}
		return nil, nil
	}

	answer := func(ctx agent.Context, input string) (string, error) {
		return "Answer: " + input, nil
	}

	state := func(ctx agent.Context, input string) (string, error) {
		return "Statement: " + input, nil
	}

	nodeConfig := workflow.NodeConfig{}
	classifyNode := workflow.NewEmittingFunctionNode("classify", classify, nodeConfig)
	answerNode := workflow.NewFunctionNode("answer", answer, nodeConfig)
	stateNode := workflow.NewFunctionNode("state", state, nodeConfig)

	edges := workflow.Concat(
		workflow.Chain(workflow.Start, classifyNode),
		[]workflow.Edge{
			{From: classifyNode, To: answerNode, Route: workflow.StringRoute("question")},
			{From: classifyNode, To: stateNode, Route: workflow.StringRoute("statement")},
		},
	)

	return workflowagent.New(workflowagent.Config{
		Name:  "routing_workflow",
		Edges: edges,
	})
}

// NewLoopWorkflow creates a loop workflow.
func NewLoopWorkflow() (agent.Agent, error) {
	type LoopState struct {
		Count int
		Max   int
	}

	initFn := func(ctx agent.Context, input string) (LoopState, error) {
		return LoopState{Count: 0, Max: 3}, nil
	}

	incrementFn := func(ctx agent.Context, state LoopState) (LoopState, error) {
		state.Count++
		return state, nil
	}

	checkFn := func(ctx agent.Context, state LoopState, emit func(*session.Event) error) (any, error) {
		ev := session.NewEvent(ctx, ctx.InvocationID())
		if state.Count >= state.Max {
			ev.Routes = []string{"done"}
		} else {
			ev.Routes = []string{"continue"}
		}
		ev.Output = state
		if err := emit(ev); err != nil {
			return nil, err
		}
		return nil, nil
	}

	formatFn := func(ctx agent.Context, state LoopState) (string, error) {
		return fmt.Sprintf("Loop finished after %d iterations!", state.Count), nil
	}

	nodeConfig := workflow.NodeConfig{}
	initNode := workflow.NewFunctionNode("init", initFn, nodeConfig)
	incrementNode := workflow.NewFunctionNode("increment", incrementFn, nodeConfig)
	checkNode := workflow.NewEmittingFunctionNode("check", checkFn, nodeConfig)
	formatNode := workflow.NewFunctionNode("format", formatFn, nodeConfig)

	eb := workflow.NewEdgeBuilder()
	eb.Add(workflow.Start, initNode)
	eb.Add(initNode, incrementNode)
	eb.Add(incrementNode, checkNode)
	eb.AddRoute(checkNode, incrementNode, workflow.StringRoute("continue"))
	eb.AddRoute(checkNode, formatNode, workflow.StringRoute("done"))

	return workflowagent.New(workflowagent.Config{
		Name:  "loop_workflow",
		Edges: eb.Build(),
	})
}

// NewHITLWorkflow creates a human-in-the-loop workflow.
func NewHITLWorkflow() (agent.Agent, error) {
	rerun := true
	greet := workflow.NewEmittingFunctionNode[any, any]("greet",
		func(nc agent.Context, _ any, emit func(*session.Event) error) (any, error) {
			reply, err := workflow.ResumeOrRequestInput(nc, emit, session.RequestInput{
				InterruptID: "ask-" + nc.InvocationID(),
				Message:     "Hello?",
			})
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Received: %v", reply), nil
		},
		workflow.NodeConfig{RerunOnResume: &rerun},
	)

	return workflowagent.New(workflowagent.Config{
		Name:  "hitl_workflow",
		Edges: workflow.Chain(workflow.Start, greet),
	})
}
