package generic

import (
	"context"
	"encoding/json"

	"github.com/Swarmind/libagent/internal/tools"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Agent struct {
	LLM           *openai.LLM
	ToolsExecutor *tools.ToolsExecutor

	toolsList *[]llms.Tool
}

func (a *Agent) Run(
	ctx context.Context,
	state []llms.MessageContent,
) (llms.MessageContent, error) {
	if a.toolsList == nil {
		a.toolsList = &[]llms.Tool{}
	}
	if len(*a.toolsList) == 0 {
		*a.toolsList = a.ToolsExecutor.ToolsList()
	}

	response, err := a.LLM.GenerateContent(
		ctx, state, llms.WithTools(*a.toolsList),
	)
	if err != nil {
		return llms.MessageContent{}, err
	}

	content := response.Choices[0].Content
	if toolContent := a.ToolsExecutor.ProcessToolCalls(
		ctx, response.Choices[0].ToolCalls,
	); toolContent != "" {
		content = toolContent
	}

	return llms.TextParts(llms.ChatMessageTypeAI, content), nil
}

func (a *Agent) SimpleRun(
	ctx context.Context,
	input string,
) (string, error) {
	if a.toolsList == nil {
		a.toolsList = &[]llms.Tool{}
	}
	if len(*a.toolsList) == 0 {
		*a.toolsList = a.ToolsExecutor.ToolsList()
	}

	response, err := a.LLM.GenerateContent(ctx,
		[]llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman,
				input,
			)},
		llms.WithTools(*a.toolsList),
	)
	if err != nil {
		return "", err
	}

	content := response.Choices[0].Content
	if toolContent := a.ToolsExecutor.ProcessToolCalls(
		ctx, response.Choices[0].ToolCalls,
	); toolContent != "" {
		content = toolContent
	}

	jsonSafeContent, err := json.Marshal(content)
	if err != nil {
		return content, err
	}

	return string(jsonSafeContent), nil
}

func (a Agent) GetLLM() *openai.LLM {
	return a.LLM
}

func (a Agent) GetToolsExecutor() *tools.ToolsExecutor {
	return a.ToolsExecutor
}
