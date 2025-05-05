package tools

import (
	"context"
	"libagent/internal/tools"
	"libagent/pkg/config"
	"os/exec"

	"github.com/tmc/langchaingo/llms"
)

// SimpleCommandExecutor represents a tool that executes commands using exec.Command.
type SimpleCommandExecutor struct{}

// Call executes the command with the given arguments.
func (s SimpleCommandExecutor) Call(ctx context.Context, input string) (string, error) {
	// args := input
	cmd := exec.Command(input)

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
				Definition: llms.FunctionDefinition{
					Name:        "simpleCommandExecutor",
					Description: "Executes a command with provided arguments using exec.Command.",
				},
				Call: SimpleCommandExecutor{}.Call,
			}, nil
		},
	)
}
