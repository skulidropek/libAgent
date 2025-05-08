package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Swarmind/libagent/pkg/config"
	"github.com/Swarmind/libagent/pkg/tools"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
	This example shows how to manually use specific tool, without calling it through llm tool call.
*/

const Prompt = `You are a hacking assistant with access to various tools for research, such as nmap.
Given user mission - find a possible attack vector and create a plan.

User Mission: 
%s`

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("new config")
	}
	if cfg.AIURL == "" {
		log.Fatal().Err(err).Msg("empty AI URL")
	}
	if cfg.AIToken == "" {
		log.Fatal().Err(err).Msg("empty AI Token")
	}
	if cfg.Model == "" {
		log.Fatal().Err(err).Msg("empty model")
	}
	cfg.SemanticSearchDisable = true

	ctx := context.Background()

	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("new tools executor")
	}

	fmt.Println("Enter you task:")
	userMission := ""
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		userMission = scanner.Text()
	}

	rewooQuery := tools.ReWOOToolArgs{
		Query: fmt.Sprintf(Prompt, userMission),
	}
	rewooQueryBytes, err := json.Marshal(rewooQuery)
	if err != nil {
		log.Fatal().Err(err).Msg("json marhsal rewooQuery")
	}

	result, err := toolsExecutor.CallTool(ctx,
		tools.ReWOOToolDefinition.Name,
		string(rewooQueryBytes),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("rewoo tool call")
	}

	if result == "" {
		log.Fatal().Msg("main empty result")
	}

	fmt.Printf("Task:\n%s\n\nResult:\n%s\n", userMission, result)
}
