package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Swarmind/libagent/pkg/config"
	"github.com/Swarmind/libagent/pkg/tools"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const defaultRepoURL = "https://github.com/JackBekket/Reflexia"

/*
	Default repository: https://github.com/Swarmind/Reflexia
	Default semantic search collection (if default repo is used): Reflexia

	Usage:
		go run examples/codemonkey/main.go -issue="Issue description or URL" [-repo="<other_repo_url>"] [-additionalContext="<extra info>"]
*/

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Define command-line flags
	repoURL := flag.String("repo", defaultRepoURL, "Repository URL")
	issueInput := flag.String("issue", "", "Issue number, full URL, or descriptive text (required)")
	additionalCtx := flag.String("additionalContext", "", "Optional additional context or instructions for the agent about the repository or task.")
	flag.Parse()

	if *issueInput == "" { // repoURL has a default, so only issueInput needs to be checked for presence
		log.Fatal().Msg("Issue input is required. Use -issue flag.")
		return
	}

	log.Debug().Str("issue_input", *issueInput).Msg("Received issue input")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize application config.")
		return
	}

	ctx := context.Background()

	toolsToWhitelist := []string{
		tools.ReWOOToolDefinition.Name,
		tools.CommandExecutorDefinition.Name,
		tools.SemanticSearchDefinition.Name,
		tools.WebReaderDefinition.Name,
	}
	log.Debug().Interface("whitelisted_tools", toolsToWhitelist).Msg("Initializing ToolsExecutor.")

	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg, tools.WithToolsWhitelist(toolsToWhitelist...))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create tools executor.")
		return
	}
	defer func() {
		if err := toolsExecutor.Cleanup(); err != nil {
			log.Error().Err(err).Msg("Error during tools executor cleanup.")
		}
	}()

	// Derive repository name and owner/repo string
	var repoName string
	var ownerAndRepo string

	trimmedRepoURL := strings.TrimSuffix(*repoURL, ".git")
	if trimmedRepoURL == strings.TrimSuffix(defaultRepoURL, ".git") {
		repoName = "Reflexia" // Default project name for semantic search collection
		ownerAndRepo = "JackBekket/Reflexia"
	} else {
		repoName = filepath.Base(trimmedRepoURL)
		parts := strings.Split(trimmedRepoURL, "/")
		if len(parts) >= 2 {
			ownerAndRepo = fmt.Sprintf("%s/%s", parts[len(parts)-2], parts[len(parts)-1])
		} else {
			log.Warn().Str("repoURL", *repoURL).Msgf("Could not reliably determine owner/repo from repoURL. Using derived repoName '%s' for PRs if needed.", repoName)
			ownerAndRepo = repoName // Fallback, might not be correct for gh CLI
		}
	}
	log.Debug().Str("derived_repo_name", repoName).Str("derived_owner_repo", ownerAndRepo).Msg("Repository details derived.")

	additionalContextString := ""
	if *additionalCtx != "" {
		// Format the additional context nicely for the prompt
		additionalContextString = fmt.Sprintf("\n- Additional User-Provided Context: %s", *additionalCtx)
	}

	// High-level task description for ReWOO
	taskDescription := fmt.Sprintf(`
Objective: As an autonomous AI Code Monkey, your mission is to comprehensively manage a GitHub issue. This includes understanding the issue, implementing the necessary code changes within the specified repository, and submitting a complete pull request with your solution.

Key Inputs for Your Mission:
- Target Repository URL: %s
- Issue Identifier (URL or details): %s%s
- Semantic Search Collection Name (should you choose to use SemanticSearchTool): '%s'
- Repository for GitHub CLI PR creation (owner/repo format): '%s'

Expected Deliverable:
- A pull request created on GitHub for the repository '%s'. This PR should effectively resolve the identified issue.
- A final summary report detailing your overall strategy, key actions, the plan you devised and followed, and the URL of the pull request if successfully created. If insurmountable errors occurred, they should be clearly documented in this report.

Core Guidelines and Operational Notes:
- Issue Investigation: If the 'Issue Identifier' is a URL, employ the 'WebReaderTool' to fetch and understand its full content. A deep understanding of the problem is paramount.
- Repository Interaction: All Git operations (cloning, branching, checking out, adding files, committing, pushing) and GitHub CLI interactions (e.g., 'gh pr create' for pull requests) must be performed using the 'CommandExecutor' tool.
    - Note on 'CommandExecutor': This tool executes commands within an isolated temporary directory. Your generated plan must account for this by ensuring commands like 'git' or file manipulations are executed in the correct context, typically by first using 'cd ./%s' to navigate into the cloned repository's subdirectory (e.g., './%s').
- Branching Strategy: Devise and create a new branch for your work. The name should be descriptive and based on the issue.
- Code Analysis and Context (Semantic Search): For non-trivial issues requiring an understanding of existing code, using the 'SemanticSearchTool' is highly recommended. Query the specified 'Semantic Search Collection Name' ('%s') with terms derived from the issue to retrieve relevant code snippets.
- Code Implementation: Based on your analysis, plan and implement the required code modifications. Use 'CommandExecutor' for all file changes.
- Pull Request Details: The pull request title should be clear and concise. The body should provide a summary of the problem, the implemented solution, and a reference to the original issue.

Your overall plan should naturally incorporate phases such as: detailed issue comprehension, local repository setup (clone & branch), problem analysis (optionally with semantic search), code implementation, thorough committing of changes, pushing to remote, and finally, pull request submission.

You have access to a suite of tools. Formulate a robust plan and execute it. Adapt to challenges as they arise.
The system will provide you with a list of available tools.
`, *repoURL, *issueInput, additionalContextString, repoName, ownerAndRepo, *repoURL, repoName, repoName, repoName)

	log.Info().Str("repository", *repoURL).Str("issue", *issueInput).Msg("Initiating ReWOO agent for Code Monkey task.")
	if *additionalCtx != "" {
		log.Info().Str("additional_context_provided", *additionalCtx).Msg("Additional context will be used.")
	}
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		log.Debug().Str("rewoo_task_description_length", fmt.Sprintf("%d chars", len(taskDescription))).Msg("Constructed ReWOO Task for Code Monkey")
	}

	rewooQueryArgs := tools.ReWOOToolArgs{
		Query: taskDescription,
	}
	rewooQueryBytes, err := json.Marshal(rewooQueryArgs)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal ReWOO query arguments.")
		return
	}

	result, err := toolsExecutor.CallTool(ctx,
		tools.ReWOOToolDefinition.Name,
		string(rewooQueryBytes),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("ReWOO agent task failed.")
		return
	}

	fmt.Println("\n--- ReWOO Agent Task Execution Report ---")
	fmt.Println(result)
	fmt.Println("\n--- End of ReWOO Agent Task ---")
	fmt.Println("\nOperational Reminders:")
	fmt.Println("1. Ensure 'git' and GitHub CLI ('gh') are installed, configured, and available in the system PATH.")
	fmt.Println("2. 'gh' CLI must be authenticated with GitHub (e.g., via 'gh auth login') for PR creation.")
	fmt.Println("3. For 'SemanticSearchTool' to be effective, a 'pgvector' database must be set up, and the relevant collection (e.g., 'Reflexia' or your specified repository name) must be populated with code embeddings from the target repository.")
}
