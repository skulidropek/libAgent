package main

import (
	"context"
	"libagent/pkg/agent/rewoo"
	"libagent/pkg/tools"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools/duckduckgo"
)

const TestPrompt = `Using duck-duck go search tool, which can search across web, create a cyberattack vector. 
First analyse user mission and point a target. 
Then try to get some information about target by web.
Then create cyberattack vector to this target. 
Then think what kind of CVE's can be used to actual perform attack on the target
Then use search and find any relevant information about this CVE in a web
Return user this plan.

User Mission:
	
`

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	baseURL := getEnv("API_URL")
	apiToken := getEnv("API_TOKEN")
	modelName := getEnv("MODEL")
	semanticSearchDBConnection := getEnv("SEMANTIC_SEARCH_DB_CONNECTION")

	llm, err := openai.New(
		openai.WithToken(apiToken),
		openai.WithModel(modelName),
		openai.WithBaseURL(baseURL),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ddgTool, err := duckduckgo.New(5, duckduckgo.DefaultUserAgent)
	if err != nil {
		log.Fatal(err)
	}
	toolsExecutor, err := tools.NewToolsExecutor(
		&tools.SemanticSearchTool{
			AIURL:          baseURL,
			AIToken:        apiToken,
			DBConnection:   semanticSearchDBConnection,
			EmbeddingModel: "text-embedding-ada-002",
			MaxResults:     2,
		},
		ddgTool,
	)
	if err != nil {
		log.Fatal(err)
	}

	rewooAgent := rewoo.Agent{
		LLM:           llm,
		ToolsExecutor: toolsExecutor,
	}

	result, err := rewooAgent.SimpleRun(context.Background(), TestPrompt)
	if err != nil {
		log.Fatal(err)
	}

	if result == "" {
		log.Fatal("empty result")
	}

	log.Printf("Result: %s", result)

}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s env is empty", key)
	}

	return val
}
