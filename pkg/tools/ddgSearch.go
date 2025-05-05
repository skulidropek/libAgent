package tools

import (
	"context"
	"encoding/json"
	"libagent/internal/tools"
	"libagent/pkg/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools/duckduckgo"
)

var DDGSearchDefinition = llms.FunctionDefinition{
	Name: "webSearch",
	Description: `A duckduckgo search wrapper.
Given search query returns a multiple results with short descriptions and URLs.`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The duckduckgo search query",
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

func (t DDGSearchTool) Call(ctx context.Context, input string) (string, error) {
	ddgSearchArgs := DDGSearchArgs{}
	if err := json.Unmarshal([]byte(input), &ddgSearchArgs); err != nil {
		return "", err
	}
	return t.wrappedTool.Call(ctx, ddgSearchArgs.Query)
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			if cfg.DDGSearchDisable {
				return nil, nil
			}
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

			return &tools.ToolData{
				Definition: DDGSearchDefinition,
				Call:       ddgSearchTool.Call,
			}, nil
		},
	)
}
