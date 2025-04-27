package tools

import (
	"context"
	"encoding/json"
	"libagent/pkg/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools/duckduckgo"
)

var DDGSearchDefinition = llms.FunctionDefinition{
	Name: "search",
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
	globalToolsRegistry = append(globalToolsRegistry,
		func(cfg config.Config) (*ToolData, error) {
			if cfg.DDGSearchMaxResults == 0 {
				cfg.DDGSearchMaxResults = 5
			}
			if cfg.DDGSearchUserAgent == "" {
				cfg.DDGSearchUserAgent = duckduckgo.DefaultUserAgent
			}

			wrappedTool, err := duckduckgo.New(
				cfg.DDGSearchMaxResults,
				cfg.DDGSearchUserAgent,
			)
			if err != nil {
				return nil, err
			}

			ddgSearchTool := DDGSearchTool{
				wrappedTool: wrappedTool,
			}

			DDGSearchDefinition.Description = wrappedTool.Description()
			return &ToolData{
				Definition: DDGSearchDefinition,
				Call:       ddgSearchTool.Call,
			}, nil
		},
	)
}
