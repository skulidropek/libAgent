package main

import (
	"context"
	"fmt"
	"libagent/pkg/agent/generic"
	"libagent/pkg/config"
	"libagent/pkg/tools"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms/openai"
)

const Prompt = `Please use rewoo tool with the next prompt:
Using semantic search tool, which can search across various code from the project
collections find out the telegram library name in the code file contents for the project called "Hellper".
Extract it from the given code and use a web search to find the pkg.go.dev documentation for it.
Give me the URL for it.`

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config.NewConfig")
	}
	if cfg.AIURL == "" {
		log.Fatal().Err(err).Msg("main empty OpenAI URL")
	}
	if cfg.AIToken == "" {
		log.Fatal().Err(err).Msg("main empty OpenAI Token")
	}
	if cfg.Model == "" {
		log.Fatal().Err(err).Msg("main empty model")
	}

	ctx := context.Background()
	agent := generic.Agent{}

	llm, err := openai.New(
		openai.WithBaseURL(cfg.AIURL),
		openai.WithToken(cfg.AIToken),
		openai.WithModel(cfg.Model),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("openai.New")
	}
	agent.LLM = llm

	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("tools.NewToolsExecutor")
	}
	agent.ToolsExecutor = toolsExecutor

	result, err := agent.SimpleRun(ctx, Prompt)
	if err != nil {
		log.Fatal().Err(err).Msg("main rewooAgent.SimpleRun")
	}
	fmt.Println(result)
}
