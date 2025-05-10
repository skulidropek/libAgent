package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/rs/zerolog/log"

	"github.com/tmc/langchaingo/llms"
)

var CommandExecutorDefinition = llms.FunctionDefinition{
	Name: "commandExecutor",
	Description: `Executes a provided string command in the bash -c wrapper.
Uses temporary home directory, so the intermediate data can be stored freely.`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "the shell command to execute to",
			},
		},
	},
}

type CommandExecutorArgs struct {
	Command string `json:"command"`
}

// CommandExecutorTool represents a tool that executes commands using exec.Command.
type CommandExecutorTool struct {
	tempDir *string
}

// Call executes the command with the given arguments.
func (s *CommandExecutorTool) Call(ctx context.Context, input string) (string, error) {
	commandExecutorArgs := CommandExecutorArgs{}
	if err := json.Unmarshal([]byte(input), &commandExecutorArgs); err != nil {
		return "", err
	}

	if s.tempDir == nil {
		tDir, err := os.MkdirTemp("", "libagent_command_executor_session_")
		if err != nil {
			return "", err
		}
		log.Debug().Msgf("command executor temp directory %s created", tDir)
		s.tempDir = &tDir
	}

	log.Debug().Msgf("command executor: bash -c %s", commandExecutorArgs.Command)
	cmd := exec.Command("bash", "-c", commandExecutorArgs.Command)
	cmd.Dir = *s.tempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (s *CommandExecutorTool) cleanup() error {
	if s.tempDir == nil {
		return nil
	}

	log.Debug().Msgf("command executor remove temp directory %s", *s.tempDir)
	return os.RemoveAll(*s.tempDir)
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			if cfg.CommandExecutorDisable {
				return nil, nil
			}

			commandExecutorTool := CommandExecutorTool{}

			return &tools.ToolData{
				Definition: CommandExecutorDefinition,
				Call:       commandExecutorTool.Call,
				Cleanup:    commandExecutorTool.cleanup,
			}, nil
		},
	)
}
