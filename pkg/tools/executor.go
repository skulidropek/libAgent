package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"

	"github.com/tmc/langchaingo/llms"
)

var SimpleCommandExecutorDefinition = llms.FunctionDefinition{
	Name:        "simpleCommandExecutor",
	Description: "Executes a shell command with provided arguments using exec.Command.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "the shell command to execute to.",
			},
		},
	},
}

type SimpleCommandExecutorArgs struct {
	Command string `json:"command"`
}

// SimpleCommandExecutorTool represents a tool that executes commands using exec.Command.
type SimpleCommandExecutorTool struct{}

// Call executes the command with the given arguments.
func (s SimpleCommandExecutorTool) Call(ctx context.Context, input string) (string, error) {
	simpleCommandExecutorArgs := SimpleCommandExecutorArgs{}
	if err := json.Unmarshal([]byte(input), &simpleCommandExecutorArgs); err != nil {
		return "", err
	}
	args := strings.Fields(simpleCommandExecutorArgs.Command)
	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			if cfg.SimpleCMDExecutorDisable {
				return nil, nil
			}
			return &tools.ToolData{
				Definition: SimpleCommandExecutorDefinition,
				Call:       SimpleCommandExecutorTool{}.Call,
			}, nil
		},
	)
}
