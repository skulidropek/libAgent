package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
)

type ToolData struct {
	Definition llms.FunctionDefinition
	Call       func(context.Context, string) (string, error)
}

type ToolsExecutor struct {
	Tools map[string]*ToolData
}

func (e ToolsExecutor) Execute(ctx context.Context, call llms.ToolCall) (llms.ToolCallResponse, error) {
	response := llms.ToolCallResponse{
		ToolCallID: call.ID,
		Name:       call.FunctionCall.Name,
	}
	toolData, ok := e.Tools[call.FunctionCall.Name]
	if !ok {
		return response, fmt.Errorf("no such tool")
	}

	var err error
	response.Content, err = toolData.Call(ctx, call.FunctionCall.Arguments)
	return response, err
}

func (e ToolsExecutor) ToolsList() []llms.Tool {
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

func (e ToolsExecutor) ToolsPromptDesc() string {
	desc := ""

	funcDefs := []llms.FunctionDefinition{
		{
			Name: "LLM",
			Description: `A pretrained LLM like yourself. Useful when you need to act with general
world knowledge and common sense. Prioritize it when you are confident in solving the problem
yourself. Input can be any instruction or task.`,
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

func (e ToolsExecutor) ProcessToolCalls(ctx context.Context, calls []llms.ToolCall) string {
	content := ""
	for _, toolCall := range calls {
		response, err := e.Execute(ctx, toolCall)
		if err != nil {
			log.Warn().Err(err).Msgf("Tool %s call", toolCall.FunctionCall.Name)
			content = fmt.Sprintf("Error calling tool %s: %v", toolCall.FunctionCall.Name, err)
			continue
		}
		content = response.Content
	}
	return content
}
