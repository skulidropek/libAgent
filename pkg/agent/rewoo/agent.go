package rewoo

import (
	"context"
	"libagent/pkg/tools"

	graph "github.com/JackBekket/langgraphgo/graph/stategraph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	GraphPlanName  = "plan"
	GraphToolName  = "tool"
	GraphSolveName = "solve"
)

type Agent struct {
	LLM           *openai.LLM
	ToolsExecutor *tools.ToolsExectutor

	graph *graph.Runnable
}

func (a Agent) Run(
	ctx context.Context,
	state []llms.MessageContent,
) (llms.MessageContent, error) {
	// TODO integrate regular chat flow into ReWOO
	return llms.MessageContent{}, nil
}

func (a Agent) SimpleRun(
	ctx context.Context,
	input string,
) (string, error) {

	if a.graph == nil {
		var err error
		a.graph, err = a.createGraph()
		if err != nil {
			return "", err
		}
	}

	state, err := a.graph.Invoke(ctx, State{
		Task: input,
	})
	if err != nil {
		return "", err
	}

	return state.(State).Result, nil
}

func (a Agent) createGraph() (*graph.Runnable, error) {
	workflowGraph := graph.NewStateGraph()

	workflowGraph.AddNode("plan", a.GetPlan)
	workflowGraph.AddNode("tool", a.ToolExecution)
	workflowGraph.AddNode("solve", a.Solve)
	workflowGraph.AddEdge("plan", "tool")
	workflowGraph.AddEdge("solve", graph.END)
	workflowGraph.AddConditionalEdge("tool", a.Route)
	workflowGraph.SetEntryPoint("plan")

	return workflowGraph.Compile()
}
