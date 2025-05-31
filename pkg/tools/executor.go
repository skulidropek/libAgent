package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/ThomasRooney/gexpect"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/tmc/langchaingo/llms"
)

const CommandNotesPromptAddition = `List of host-specific command substitutions and their descriptions:`

var CommandExecutorDefinition = llms.FunctionDefinition{
	Name: "commandExecutor",
	Description: `Executes a provided string command in the interactive bash shell.
Uses temporary home directory.
Most likely all needed packages are preinstalled.`,
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

	process *gexpect.ExpectSubprocess
	prompt  string
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

		s.process, err = gexpect.SpawnAtDirectory("env -i bash --norc --noprofile", *s.tempDir)
		if err != nil {
			return "", fmt.Errorf("spawn: %w", err)
		}
		s.process.Capture()
		s.process.Expect("$")

		s.prompt = uuid.New().String()
		s.process.Send(fmt.Sprintf("PS1=%s\n", s.prompt))
		s.process.Expect(fmt.Sprintf("PS1=%s", s.prompt))
		s.process.Expect(s.prompt)
		s.process.Collect()
	}

	log.Debug().Msgf("command executor: %s", commandExecutorArgs.Command)
	s.process.Capture()
	s.process.Send(commandExecutorArgs.Command + "\n")
	s.process.Expect(s.prompt)

	output := ""
	outputLines := strings.Split(strings.TrimSpace(string(s.process.Collect())), "\n")
	if len(outputLines) > 2 {
		output = strings.Join(outputLines[1:len(outputLines)-1], "\n")
	}

	log.Debug().Msgf("command output: %s", output)

	return output, nil
}

func (s *CommandExecutorTool) cleanup() error {
	if s.tempDir == nil {
		return nil
	}

	log.Debug().Msgf("command executor remove temp directory and process shutdown %s", *s.tempDir)
	err := os.RemoveAll(*s.tempDir)
	s.tempDir = nil
	if err != nil {
		log.Warn().Err(err).Msg("Removing temp dir")
	}

	return s.process.Close()
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			if cfg.CommandExecutorDisable {
				return nil, nil
			}

			commandExecutorTool := CommandExecutorTool{}

			definition := CommandExecutorDefinition
			if len(cfg.CommandExecutorCommands) > 0 {
				commandsList := ""
				for command, description := range cfg.CommandExecutorCommands {
					descSplit := strings.SplitN(description, "\n", 1)
					if len(descSplit) == 2 {
						commandsList += fmt.Sprintf(
							"- %s: %s\n%s\n",
							strings.ToLower(command),
							descSplit[0], descSplit[1],
						)
					} else {
						commandsList += fmt.Sprintf(
							"- %s: %s\n",
							strings.ToLower(command),
							description,
						)
					}
				}

				definition.Description += fmt.Sprintf(
					"\n%s\n%s",
					CommandNotesPromptAddition, commandsList,
				)
			}

			if strings.HasSuffix(definition.Description, "\n\n") {
				definition.Description = strings.TrimSuffix(definition.Description, "\n")
			}

			return &tools.ToolData{
				Definition: definition,
				Call:       commandExecutorTool.Call,
				Cleanup:    commandExecutorTool.cleanup,
			}, nil
		},
	)
}
