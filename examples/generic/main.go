package main

import (
	"context"
	"fmt"
	"libagent/pkg/agent/rewoo"
	"libagent/pkg/config"
	"libagent/pkg/tools"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms/openai"
)

const InnerPrompt = `Using semantic search tool, which can search across various code from the project
collections find out the telegram library name in the code file contents for the project called "Hellper".
Extract it from the given code and use a web search to find the pkg.go.dev documentation for it.
Give me the URL for it.`

const Prompt = `This is a test of your ability to use various tools.
I need you to simply call the tool called "rewoo" for creating subtasks with the next prompt and return the result.
Prompt:
` + InnerPrompt

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
	rewooAgent := rewoo.Agent{}
	ctx := context.WithValue(context.Background(), "ReWOOAgent", &rewooAgent)

	llm, err := openai.New(
		openai.WithBaseURL(cfg.AIURL),
		openai.WithToken(cfg.AIToken),
		openai.WithModel(cfg.Model),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("openai.New")
	}
	rewooAgent.LLM = llm

	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("tools.NewToolsExecutor")
	}
	rewooAgent.ToolsExecutor = toolsExecutor

	result, err := rewooAgent.SimpleRun(ctx, Prompt)
	if err != nil {
		log.Fatal().Err(err).Msg("main rewooAgent.SimpleRun")
	}

	if result == "" {
		log.Fatal().Msg("main empty result")
	}

	fmt.Println(result)
}
