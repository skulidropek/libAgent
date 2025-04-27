package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"libagent/pkg/config"
	"slices"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

const LLMToolName = "LLM"

type ToolData struct {
	Definition llms.FunctionDefinition
	Call       func(context.Context, string) (string, error)
}

type ToolsExectutor struct {
	Tools map[string]*ToolData
}

var globalToolsRegistry = []func(config.Config) (*ToolData, error){}

var ErrNoSuchTool = errors.New("no such tool")

func NewToolsExecutor(cfg config.Config) (*ToolsExectutor, error) {
	toolsExecutor := ToolsExectutor{}
	tools := map[string]*ToolData{}
	for _, toolInit := range globalToolsRegistry {
		tool, err := toolInit(cfg)
		if err != nil {
			return nil, err
		}
		tools[tool.Definition.Name] = tool
	}
	toolsExecutor.Tools = tools

	return &toolsExecutor, nil
}

func (e ToolsExectutor) Execute(ctx context.Context, call llms.ToolCall) (llms.ToolCallResponse, error) {
	response := llms.ToolCallResponse{
		ToolCallID: call.ID,
		Name:       call.FunctionCall.Name,
	}
	toolData, ok := e.Tools[call.FunctionCall.Name]
	if !ok {
		return response, ErrNoSuchTool
	}

	var err error
	response.Content, err = toolData.Call(ctx, call.FunctionCall.Arguments)
	return response, err
}

func (e ToolsExectutor) ToolsList() []llms.Tool {
	tools := []llms.Tool{}
	for _, toolData := range e.Tools {
		tools = append(tools, llms.Tool{
			Type:     "function",
			Function: &toolData.Definition,
		})
	}
	slices.SortFunc(tools,
		func(a, b llms.Tool) int {
			return strings.Compare(a.Function.Name, b.Function.Name)
		},
	)

	return tools
}

func (e ToolsExectutor) ToolsPromptDesc() string {
	desc := ""

	funcDefs := []llms.FunctionDefinition{
		{
			Name: LLMToolName,
			Description: `A pretrained LLM like yourself. Useful when you need to act with general
world knowledge and common sense. Prioritize it when you are confident in solving the problem
yourself. Input can be any instruction.`,
		},
	}
	for _, toolData := range e.Tools {
		funcDefs = append(funcDefs, toolData.Definition)
	}
	slices.SortFunc(funcDefs,
		func(a, b llms.FunctionDefinition) int {
			return strings.Compare(a.Name, b.Name)
		},
	)

	for idx, def := range funcDefs {
		input := "string"
		if def.Parameters != nil {
			if props, ok := def.Parameters.(map[string]interface{})["properties"]; ok {
				fields := map[string]string{}
				for prop, val := range props.(map[string]interface{}) {
					fields[prop] = val.(map[string]interface{})["type"].(string)
				}
				fieldsJson, err := json.Marshal(fields)
				if err == nil {
					input = string(fieldsJson)
				}
			}
		}
		desc += fmt.Sprintf("(%d) %s[%s]: %s\n", idx, def.Name, input, def.Description)
	}
	return desc
}
