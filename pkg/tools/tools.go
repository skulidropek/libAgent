package tools

import (
	"context"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"
)

var globalToolsRegistry = []func(context.Context, config.Config) (*tools.ToolData, error){}
var globalToolsExecutor *tools.ToolsExecutor

func NewToolsExecutor(ctx context.Context, cfg config.Config) (*tools.ToolsExecutor, error) {
	toolsExecutor := tools.ToolsExecutor{}
	tools := map[string]*tools.ToolData{}
	for _, toolInit := range globalToolsRegistry {
		tool, err := toolInit(ctx, cfg)
		if err != nil {
			return nil, err
		}
		if tool == nil {
			continue
		}
		tools[tool.Definition.Name] = tool
	}
	toolsExecutor.Tools = tools

	if globalToolsExecutor == nil {
		globalToolsExecutor = &toolsExecutor
	}

	return &toolsExecutor, nil
}
