package agent

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

type Agent interface {
	Run(
		ctx context.Context,
		state []llms.MessageContent,
	) (llms.MessageContent, error)
	SimpleRun(
		ctx context.Context,
		input string,
	) (string, error)
}
