package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Swarmind/libagent/pkg/config"
	"github.com/Swarmind/libagent/pkg/tools"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
	This example demonstrates how to use the Code Monkey Agent to automatically
	resolve issues in a codebase and create pull requests with the fixes.

	Usage:
		go run main.go -repo="https://github.com/owner/repo.git" -issue="123"
		or
		go run main.go -repo="https://github.com/StarumDDD/Personalized-portfolio-cite-TS" -issue="https://github.com/StarumDDD/Personalized-portfolio-cite-TS/issues/1"
*/

func main() {
	// Set up logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Parse command line arguments
	repoURL := flag.String("repo", "", "Repository URL (required)")
	issueInput := flag.String("issue", "", "Issue number or URL (required)")
	flag.Parse()

	if *repoURL == "" || *issueInput == "" {
		log.Fatal().Msg("Repository URL and issue are required. Use -repo and -issue flags")
	}

	// Initialize config
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
	if cfg.GitHubToken == "" {
		log.Fatal().Err(err).Msg("empty GitHub Token")
	}

	// Create context
	ctx := context.Background()

	// Create tools executor with specific tools
	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg, tools.WithToolsWhitelist(
		tools.ReWOOToolDefinition.Name,
		tools.CommandExecutorDefinition.Name,
		tools.WebReaderDefinition.Name,
	))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create tools executor")
	}
	defer func() {
		if err := toolsExecutor.Cleanup(); err != nil {
			log.Fatal().Err(err).Msg("tools executor cleanup")
		}
	}()

	// Create prompt for ReWOO
	prompt := fmt.Sprintf(`You will be provided with tool name and arguments for the call.
Use tool description and plan to decide how to resolve the provided arguments into the tool call schema.
Try to sanitize arguments, resolve possible string concatenation.

Plan: 
1. Use webReader to analyze GitHub repository structure
2. Get issue details and repository contents
3. Analyze the code and make necessary changes
4. Create PR with the changes

First, use webReader to analyze the repository:
Tool name: webReader
Tool description: Uses provided valid URL and provides a markdown text converted from html for ease of read.
Arguments: {
  "url": "%s"
}

Then, get issue details and repository contents:
Tool name: commandExecutor
Tool description: Executes a provided string command in the bash -c wrapper.
Arguments: {
  "command": "cd \"$(pwd)\" && rm -rf repo && mkdir repo && cd repo && \
    git init && git remote add origin %s && \
    git fetch origin && git checkout -b main origin/main && \
    git config user.email \"agent@swarmind.com\" && \
    git config user.name \"CodeMonkey Agent\" && \
    git config credential.helper store && \
    mkdir -p ~/.config/gh && \
    echo \"github.com:\" > ~/.config/gh/hosts.yml && \
    echo \"  user: agent@swarmind.com\" >> ~/.config/gh/hosts.yml && \
    echo \"  oauth_token: %s\" >> ~/.config/gh/hosts.yml && \
    ISSUE_TITLE=$(gh issue view %s --json title --jq .title) && \
    ISSUE_BODY=$(gh issue view %s --json body --jq .body) && \
    echo \"Issue Title: $ISSUE_TITLE\" && \
    echo \"Issue Description: $ISSUE_BODY\" && \
    echo \"Repository Contents:\" && \
    find . -type f -not -path \"./.*\" -not -path \"./node_modules/*\" -not -path \"./dist/*\" -not -path \"./build/*\" | while read file; do \
      echo \"\n=== $file ===\" && \
      cat \"$file\" && \
      echo \"\n\"; \
    done"
}

Based on the analysis, make the necessary changes to the files:
Tool name: commandExecutor
Tool description: Executes a provided string command in the bash -c wrapper.
Arguments: {
  "command": "cd \"$(pwd)/repo\" && \
    ISSUE_TITLE=$(gh issue view %s --json title --jq .title) && \
    ISSUE_BODY=$(gh issue view %s --json body --jq .body) && \
    echo \"Making changes based on issue: $ISSUE_TITLE\" && \
    echo \"$ISSUE_BODY\" && \
    find . -type f -not -path \"./.*\" -not -path \"./node_modules/*\" -not -path \"./dist/*\" -not -path \"./build/*\" | while read file; do \
      echo \"Analyzing $file...\" && \
      if grep -q \"button\\|btn\" \"$file\"; then \
        echo \"Found button styling in $file\" && \
        if grep -q \"blue\" \"$file\"; then \
          echo \"Modifying $file to change blue to green\" && \
          sed -i 's/blue/green/g' \"$file\" && \
          echo \"Changes made to $file\" && \
        fi; \
      fi; \
    done && \
    git checkout -b monkeyagent_attempt && \
    git add . && \
    git commit -m \"Fix: $ISSUE_TITLE\" && \
    git remote set-url origin https://%s@github.com/%s && \
    git push -u origin monkeyagent_attempt && \
    gh pr create --title \"Fix: $ISSUE_TITLE\" --body \"Fixes issue: $ISSUE_TITLE\n\n$ISSUE_BODY\" --base main"
}`, *repoURL, *repoURL, cfg.GitHubToken, *issueInput, *issueInput, *issueInput, *issueInput, cfg.GitHubToken,
		strings.TrimPrefix(strings.TrimSuffix(*repoURL, ".git"), "https://github.com/"))

	// Call ReWOO tool with max length check
	if len(prompt) > 4000 {
		log.Fatal().Msg("Prompt too long, please simplify the task")
	}

	// Call ReWOO tool
	rewooQuery := tools.ReWOOToolArgs{
		Query: prompt,
	}
	rewooQueryBytes, err := json.Marshal(rewooQuery)
	if err != nil {
		log.Fatal().Err(err).Msg("json marshal rewooQuery")
	}

	result, err := toolsExecutor.CallTool(ctx,
		tools.ReWOOToolDefinition.Name,
		string(rewooQueryBytes),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("rewoo tool call")
	}

	fmt.Println("\nResults:")
	fmt.Println(result)
	fmt.Println("\nDebug Info:")
	fmt.Println("1. Make sure you have git installed and configured")
	fmt.Println("2. The repository will be cloned in the temp directory")
	fmt.Println("3. All git operations will be performed in the cloned repository")
	fmt.Println("4. Each command is chained to ensure proper directory context")
}
