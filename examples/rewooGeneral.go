package main

import (
	"context"
	"libagent/pkg/agent/rewoo"
	"libagent/pkg/tools"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
)

const TestPrompt = `Using semantic search tool, which can search across various code from the project collections find out the telegram library name in the code file contents for the project called "Hellper". Extract it from the given code and use a web search to find the pkg.go.dev documentation for it. Give me the URL for it.`

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	baseURL := getEnv("API_URL")
	apiToken := getEnv("API_TOKEN")
	modelName := getEnv("MODEL")

	llm, err := openai.New(
		openai.WithToken(apiToken),
		openai.WithModel(modelName),
		openai.WithBaseURL(baseURL),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal(err)
	}

	toolsExecutor, err := tools.NewToolsExecutor()
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
