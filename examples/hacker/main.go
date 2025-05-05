package main

import (
	"bufio"
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

const Prompt = `You are a hacking assistant with access to various tools for research.
Given user mission - find a possible attack vector and create a plan.

User Mission: 
%s`

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

	cfg.SemanticSearchDisable = true
	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("tools.NewToolsExecutor")
	}
	rewooAgent.ToolsExecutor = toolsExecutor

	fmt.Println("Enter you task:")
	userMission := ""
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		userMission = scanner.Text()
	}

	result, err := rewooAgent.SimpleRun(ctx, fmt.Sprintf(Prompt, userMission))
	if err != nil {
		log.Fatal().Err(err).Msg("main rewooAgent.SimpleRun")
	}

	if result == "" {
		log.Fatal().Msg("main empty result")
	}

	fmt.Printf("Task:\n%s\n\nResult:\n%s\n", userMission, result)
}
