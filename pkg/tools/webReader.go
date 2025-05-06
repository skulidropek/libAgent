package tools

import (
	"context"
	"encoding/json"

	"github.com/Swarmind/libagent/internal/tools"
	webreader "github.com/Swarmind/libagent/internal/tools/webReader"
	"github.com/Swarmind/libagent/pkg/config"

	"github.com/tmc/langchaingo/llms"
)

var WebReaderDefinition = llms.FunctionDefinition{
	Name: "webReader",
	Description: `Uses provided valid URL and provides a markdown text converted from html for ease of read.
		Please be sure to put a valid URL here, you can use LLM tool to extract it from query before using it in this tool.`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The valid url to read as text from.",
			},
		},
	},
}

type WebReaderArgs struct {
	URL string `json:"url"`
}

type WebReaderTool struct {
}

func (t WebReaderTool) Call(ctx context.Context, input string) (string, error) {
	webReaderArgs := WebReaderArgs{}
	if err := json.Unmarshal([]byte(input), &webReaderArgs); err != nil {
		return "", err
	}

	return webreader.ProcessUrl(webReaderArgs.URL)
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			if cfg.WebReaderDisable {
				return nil, nil
			}
			webReaderTool := WebReaderTool{}

			return &tools.ToolData{
				Definition: WebReaderDefinition,
				Call:       webReaderTool.Call,
			}, nil
		},
	)
}
