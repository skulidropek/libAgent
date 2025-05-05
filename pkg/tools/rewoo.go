package tools

import (
	"context"
	"encoding/json"
	"libagent/pkg/agent"
	"libagent/pkg/config"

	"github.com/tmc/langchaingo/llms"
)

var ReWOOToolDefinition = llms.FunctionDefinition{
	Name: "rewoo",
	Description: `A more complex LLM Reasoning algorithm.
		Useful when you need to do a tool assisted reasoning research.
		Usually tends to return a short response as a result of multiple step thinking.
		Use it is you think that you have an isolated complex research subtask.
		Input can be any complex task.`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The task query",
			},
		},
	},
}

type ReWOOToolArgs struct {
	Query string `json:"query"`
}

type ReWOOTool struct {
	agent agent.Agent
}

func (s ReWOOTool) Call(ctx context.Context, input string) (string, error) {
	rewooToolArgs := ReWOOToolArgs{}
	if err := json.Unmarshal([]byte(input), &rewooToolArgs); err != nil {
		return "", err
	}
	return s.agent.SimpleRun(ctx, rewooToolArgs.Query)
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*ToolData, error) {
			rewooTool := ReWOOTool{
				agent: ctx.Value("ReWOOAgent").(agent.Agent),
			}

			return &ToolData{
				Definition: ReWOOToolDefinition,
				Call:       rewooTool.Call,
			}, nil
		},
	)
}
