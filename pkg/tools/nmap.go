package tools

import (
	"context"
	"libagent/internal/tools"
	"libagent/pkg/config"
	"os/exec"

	"github.com/tmc/langchaingo/llms"
)

type Nmap struct{}

// Call executes the command with the given arguments.
func (s Nmap) Call(ctx context.Context, input string) (string, error) {
	args := input
	cmd := exec.Command("nmap -v -T4 -PA -sV --version-all --osscan-guess -A -sS -p 1-65535", args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			return &tools.ToolData{
				Definition: llms.FunctionDefinition{
					Name:        "Nmap",
					Description: "Executes nmap (scanning ports of) target address",
				},
				Call: Nmap{}.Call,
			}, nil
		},
	)
}
