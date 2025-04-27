package tools

import (
	"context"
	"encoding/json"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools/duckduckgo"
)

var DDGSearchDefinition = llms.FunctionDefinition{
	Name: "search",
	Description: `
	"A wrapper around DuckDuckGo Search."
	"Free search alternative to google and serpapi."
	"Input should be a search query."`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
		},
	},
}

type DDGSearchArgs struct {
	Query string `json:"query"`
}

type DDGSearchTool struct {
	wrappedTool *duckduckgo.Tool
}

func (s DDGSearchTool) Call(ctx context.Context, input string) (string, error) {
	ddgSearchArgs := DDGSearchArgs{}
	if err := json.Unmarshal([]byte(input), &ddgSearchArgs); err != nil {
		return "", err
	}
	return s.wrappedTool.Call(ctx, ddgSearchArgs.Query)
}

func init() {
	toolsRegistry = append(toolsRegistry,
		func() (ToolData, error) {
			wrappedTool, err := duckduckgo.New(5, duckduckgo.DefaultUserAgent)
			if err != nil {
				return ToolData{}, err
			}
			ddgSearchTool := DDGSearchTool{
				wrappedTool: wrappedTool,
			}

			return ToolData{
				Definition: DDGSearchDefinition,
				Call:       ddgSearchTool.Call,
			}, nil
		},
	)
}
