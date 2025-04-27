package generic

import (
	"context"
	"encoding/json"
	"libagent/pkg/tools"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Agent struct {
	LLM           *openai.LLM
	ToolsExecutor *tools.ToolsExectutor

	toolsList *[]llms.Tool
}

func (a *Agent) Run(
	ctx context.Context,
	state []llms.MessageContent,
) (llms.MessageContent, error) {
	// TODO integrate regular chat flow into generic agent
	// this is messageGraph agent, not stateGraph agent as reasoning without obsrvation agent
	return llms.MessageContent{}, nil
}

func (a *Agent) SimpleRun(
	ctx context.Context,
	input string,
) (string, error) {

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
	for _, toolCall := range response.Choices[0].ToolCalls {
		response, err := a.ToolsExecutor.Execute(ctx, toolCall)
		if err != nil {
			return "", err
		}
		content = response.Content
	}

	jsonSafeContent, err := json.Marshal(content)
	if err != nil {
		return content, err
	}

	return string(jsonSafeContent), nil
}
