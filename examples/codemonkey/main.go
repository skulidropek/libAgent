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
	This example uses the ReWOO agent to manage a GitHub issue end-to-end,
	inspired by the prompting style of examples/generic/main.go.
	The prompt guides ReWOO on tool usage for various sub-tasks.

	Default repository: https://github.com/JackBekket/Reflexia
	Default semantic search collection (if default repo is used): Reflexia
*/

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	repoURL := flag.String("repo", defaultRepoURL, "Repository URL")
	issueInput := flag.String("issue", "", "Issue number, full URL, or descriptive text (required)")
	additionalCtx := flag.String("additionalContext", "", "Optional additional context for the agent.")
	flag.Parse()

	if *issueInput == "" {
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

	var repoName string
	var ownerAndRepo string
	trimmedRepoURL := strings.TrimSuffix(*repoURL, ".git")

	if trimmedRepoURL == strings.TrimSuffix(defaultRepoURL, ".git") {
		repoName = "Reflexia"
		ownerAndRepo = "JackBekket/Reflexia"
	} else {
		repoName = filepath.Base(trimmedRepoURL)
		parts := strings.Split(trimmedRepoURL, "/")
		if len(parts) >= 2 {
			ownerAndRepo = fmt.Sprintf("%s/%s", parts[len(parts)-2], parts[len(parts)-1])
		} else {
			ownerAndRepo = repoName // Fallback
		}
	}
	log.Debug().Str("derived_repo_name", repoName).Str("derived_owner_repo", ownerAndRepo).Msg("Repository details derived.")

	additionalContextString := ""
	if *additionalCtx != "" {
		additionalContextString = fmt.Sprintf("\n\nAdditional Context Provided by User:\n%s", *additionalCtx)
	}

	// Task description guiding ReWOO, similar in style to examples/generic/main.go's prompt
	taskDescription := fmt.Sprintf(`
Your task is to act as an AI Code Monkey to resolve the GitHub issue identified by "%s" for the repository %s.
The ultimate goal is to create a pull request with the implemented fix.
%s

Please follow this general approach, utilizing the available tools as specified:

1.  **Understand the Issue**:
    * If the issue identifier ("%s") is a URL, use the 'WebReaderTool' to fetch its full content.
    * Thoroughly analyze the complete issue description to understand the problem.

2.  **Set up the Repository Environment**:
    * First, use 'CommandExecutor' to clone the repository %s into a subdirectory named '%s'.
    * Next, devise a suitable new branch name (e.g., "fix/issue-summary-slug") based on the issue's content.
    * Then, to perform subsequent Git operations like creating a branch, YOU MUST USE 'CommandExecutor' by chaining commands to ensure they run inside the cloned repository. For example, to checkout main and create your new branch, the command should be: 'cd ./%s && git checkout main && git checkout -b your-new-branch-name'.

3.  **Gather Code Context (if needed for the fix)**:
    * To understand the existing codebase relevant to the issue, use the 'SemanticSearchTool'.
    * The 'collection' for this search must be '%s'.
    * The 'query' should be derived from the detailed issue description you obtained in step 1.

4.  **Develop and Implement Code Changes**:
    * Based on your analysis of the issue and any code context from semantic search, determine the necessary code modifications.
    * If you need to generate code snippets or reason about complex changes, you can use the 'LLM' tool (which is your own reasoning capability).
    * To apply the changes to files, use 'CommandExecutor'. Remember to chain 'cd ./%s' with your file modification commands (e.g., 'cd ./%s && echo "new content" > path/to/file.go' or 'cd ./%s && sed -i "s/old/new/" path/to/file.go'). For multiple changes, consider planning a script to be written and then executed by 'CommandExecutor' within the repo directory.

5.  **Commit and Push Changes**:
    * Using 'CommandExecutor' (and adhering to the 'cd ./%s && ...' pattern), stage all your changes ('git add .').
    * Then, commit the changes with a descriptive message that references the issue.
    * Finally, push the new branch to the origin remote.

6.  **Create Pull Request**:
    * Use 'CommandExecutor' with the GitHub CLI ('gh') to create a pull request.
    * The command should be structured like: 'cd ./%s && gh pr create --repo %s --title "Fix: [Issue Title]" --body "Fixes issue: %s. [Your summary of changes]" --base [main_or_master_branch] --head your-new-branch-name'.

After completing these steps, provide a final summary report of your actions and the URL of the created pull request if successful. If any step fails, report the error clearly.
`, *issueInput, *repoURL, additionalContextString, *issueInput, *repoURL, repoName, repoName, repoName, repoName, repoName, repoName, repoName, ownerAndRepo, *issueInput)

	log.Info().Str("repository", *repoURL).Str("issue", *issueInput).Msg("Initiating ReWOO agent for Code Monkey task.")
	if *additionalCtx != "" {
		log.Info().Str("additional_context_provided_length", fmt.Sprintf("%d chars", len(*additionalCtx))).Msg("Additional context will be used.")
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
	fmt.Println("\nOperational Reminders:") // Keep these reminders for the user
	fmt.Println("1. Ensure 'git' and GitHub CLI ('gh') are installed, configured, and in system PATH.")
	fmt.Println("2. 'gh' CLI must be authenticated with GitHub (e.g., via 'gh auth login') for PR creation.")
	fmt.Println("3. For 'SemanticSearchTool', a 'pgvector' database with a collection named after the repository (e.g., 'Reflexia') must be populated with relevant code embeddings.")
}
