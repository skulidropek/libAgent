package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Swarmind/libagent/pkg/config"
	_ "github.com/Swarmind/libagent/pkg/logging"
	"github.com/Swarmind/libagent/pkg/tools"

	"github.com/rs/zerolog/log"
)

/*
	This example shows how to manually use specific tool, without calling it through llm tool call.
*/

const Prompt = `You are a hacking assistant with access to various tools for research.
Given user mission - find a possible attack vector and create a plan.

User Mission: 
%s`

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("new config")
	}
	cfg.SemanticSearchDisable = true

	ctx := context.Background()

	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("new tools executor")
	}
	defer func() {
		if err := toolsExecutor.Cleanup(); err != nil {
			log.Fatal().Err(err).Msg("tools executor cleanup")
		}
	}()

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
