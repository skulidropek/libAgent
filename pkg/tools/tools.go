package tools

import (
	"context"
	"slices"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"
)

type ExecutorOption func(*ExecutorOptions)

type ExecutorOptions struct {
	ToolsWhitelist []string
}

var globalToolsRegistry = []func(context.Context, config.Config) (*tools.ToolData, error){}
var globalToolsExecutor *tools.ToolsExecutor

func NewToolsExecutor(ctx context.Context, cfg config.Config, opts ...ExecutorOption) (*tools.ToolsExecutor, error) {
	toolsExecutor := tools.ToolsExecutor{}
	tools := map[string]*tools.ToolData{}
	options := ExecutorOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	for _, toolInit := range globalToolsRegistry {
		tool, err := toolInit(ctx, cfg)
		if err != nil {
			return nil, err
		}
		if tool == nil {
			continue
		}

		if len(options.ToolsWhitelist) > 0 &&
			!slices.Contains(
				options.ToolsWhitelist,
				tool.Definition.Name,
			) {
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

func WithToolsWhitelist(tool ...string) ExecutorOption {
	return func(eo *ExecutorOptions) {
		eo.ToolsWhitelist = append(eo.ToolsWhitelist, tool...)
	}
}
