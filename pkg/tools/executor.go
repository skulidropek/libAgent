package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/ThomasRooney/gexpect"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/tmc/langchaingo/llms"
)

const CommandNotesPromptAddition = `Important! List of host machine specific command usage recommendations:`

var CommandExecutorDefinition = llms.FunctionDefinition{
	Name: "commandExecutor",
	Description: `Executes a provided command in a interactive stateful bash shell session.
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

	return s.RunCommand(commandExecutorArgs.Command)
}

func (s *CommandExecutorTool) RunCommand(input string) (string, error) {
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
		// Start recording buffer
		s.process.Capture()
		// Expect default bash shell prompt end
		if err := s.process.Expect("$"); err != nil {
			return "", fmt.Errorf("expect initial prompt: %w", err)
		}

		if err := s.process.Send(fmt.Sprintf("PATH=%s\n", os.Getenv("PATH"))); err != nil {
			return "", fmt.Errorf("set PATH: %w", err)
		}
		// Create a random UUID to set as a prompt to be sure that there are command end
		s.prompt = uuid.New().String()
		if err := s.process.Send(fmt.Sprintf("PS1=%s\n", s.prompt)); err != nil {
			return "", fmt.Errorf("set prompt: %w", err)
		}
		// Expect sent command
		if err := s.process.Expect(fmt.Sprintf("PS1=%s", s.prompt)); err != nil {
			return "", fmt.Errorf("expect prompt confirmation: %w", err)
		}
		// Expect changed prompt
		if err := s.process.Expect(s.prompt); err != nil {
			return "", fmt.Errorf("expect prompt change: %w", err)
		}
		// Discard output by draining output buffer
		s.process.Collect()
	}

	// Trim trailing '\' to avoid escaping last '\n' symbol
	command := strings.TrimSuffix(
		strings.ReplaceAll(
			// Replace '\"' with '"' to avoid double escaping, since the command is from json payload
			strings.TrimSpace(input),
			`\"`, `"`,
		), `\`,
	) + "\n"
	log.Debug().Msgf("command executor: %s", strings.TrimSpace(command))
	s.process.Capture()
	if err := s.process.Send(command); err != nil {
		return "", fmt.Errorf("send command: %w", err)
	}
	err := s.process.ExpectTimeout(s.prompt, time.Second*30)
	if err != nil {
		if sendErr := s.process.Send(string([]byte{0x03})); sendErr != nil {
			log.Warn().Err(sendErr).Msg("command executor interrupt send")
		}
	}

	output := string(s.process.Collect())
	// Collected output will include prompt and entered command line
	// Strip the first line and trim prompt suffix (since output command can be terminated without newline)
	outputLines := strings.Split(strings.TrimSpace(output), "\n")[1:]
	output = strings.TrimSpace(strings.TrimSuffix(strings.Join(outputLines, "\n"), s.prompt))

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
