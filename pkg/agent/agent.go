package agent

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

type Agent interface {
	Run(
		ctx context.Context,
		state []llms.MessageContent,
		opts ...llms.CallOption,
	) (llms.MessageContent, error)
	SimpleRun(
		ctx context.Context,
		input string,
		opts ...llms.CallOption,
	) (string, error)
}
