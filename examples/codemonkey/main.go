package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

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
	prompt := fmt.Sprintf(`Here is the step by step actions plan to fix the issue in the repository:
1. Set up the repository and git:
   cd "$(pwd)" && \
   rm -rf repo && \
   mkdir repo && \
   cd repo && \
   git init && \
   git remote add origin %s && \
   git fetch origin && \
   git checkout -b main origin/main && \
   git config user.email "agent@swarmind.com" && \
   git config user.name "CodeMonkey Agent"

2. Read the issue %s to understand what needs to be fixed

3. Analyze the codebase:
   cd "$(pwd)/repo" && \
   find . -name "*.ts" -o -name "*.js" -o -name "*.html"

4. Make the necessary changes to fix the issue:
   cd "$(pwd)/repo" && \
   find . -type f -exec sed -i 's/blue/green/g' {} +

5. Create a new branch and commit changes:
   cd "$(pwd)/repo" && \
   git checkout -b fix-issue-1 && \
   git add . && \
   git commit -m "Fix issue #1"

6. Push changes and create a pull request:
   cd "$(pwd)/repo" && \
   git push -u origin fix-issue-1 && \
   gh pr create --title "Fix issue #1" --body "Fixes the issue as described in #1"

You need to execute every step separately and write me a report as a list like this:
- setup repository [cd "$(pwd)" && mkdir repo ...]: OK
- read issue [gh issue view ...]: <issue content>
- analyze codebase [cd "$(pwd)/repo" && find ...]: <list of files to modify>
- make changes [cd "$(pwd)/repo" && sed ...]: OK
- create branch and commit [cd "$(pwd)/repo" && git checkout ...]: OK
- push and create PR [cd "$(pwd)/repo" && git push ...]: <PR URL>

Note: Each command must be executed in the correct directory. The commands will use the current working directory.
`, *repoURL, *issueInput)

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
