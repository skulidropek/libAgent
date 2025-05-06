package tools

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/Swarmind/libagent/internal/tools"
	"github.com/Swarmind/libagent/pkg/config"

	"github.com/tmc/langchaingo/llms"
)

var NmapToolDefinition = llms.FunctionDefinition{
	Name:        "nmap",
	Description: "Executes nmap -v -T4 -PA -sV --version-all --osscan-guess -A -sS -p 1-65535 [IP], where IP is the call argument.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ip": map[string]any{
				"type":        "string",
				"description": "The valid IP address to scan to.",
			},
		},
	},
}

type NmapTool struct{}

type NmapToolArgs struct {
	IP string `json:"ip"`
}

// Call executes the command with the given arguments.
func (s NmapTool) Call(ctx context.Context, input string) (string, error) {
	nmapToolArgs := NmapToolArgs{}
	if err := json.Unmarshal([]byte(input), &nmapToolArgs); err != nil {
		return "", err
	}
	cmd := exec.Command("nmap", "-v", "-T4", "-PA", "-sV", "--version-all", "-osscan-guess", "-A", "-sS", "-p", "1-65535", nmapToolArgs.IP)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*tools.ToolData, error) {
			if cfg.NmapDisable {
				return nil, nil
			}

			return &tools.ToolData{
				Definition: NmapToolDefinition,
				Call:       NmapTool{}.Call,
			}, nil
		},
	)
}
