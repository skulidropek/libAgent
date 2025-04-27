package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

var SemanticSearchDefinition = llms.FunctionDefinition{
	Name:        "semanticSearch",
	Description: "Performs semantic search in the vector store of the saved code blobs. Returns matching file contents",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
			"collection": map[string]any{ //TODO: there should NOT exist arguments which called NAME cause it cause COLLISION with actual function name.    .....more like confusion then collision so there are no error
				"type":        "string",
				"description": "name of collection store in which we perform the search",
			},
		},
	},
}

type SemanticSearchArgs struct {
	Query      string `json:"query"`
	Collection string `json:"collection"`
}

type SemanticSearchTool struct {
	AIURL          string
	AIToken        string
	DBConnection   string
	EmbeddingModel string
	MaxResults     int
}

func (s SemanticSearchTool) Call(ctx context.Context, input string) (string, error) {
	semanticSearchArgs := SemanticSearchArgs{}
	response := ""

	if err := json.Unmarshal([]byte(input), &semanticSearchArgs); err != nil {
		return response, err
	}

	config, err := pgxpool.ParseConfig(s.DBConnection)
	if err != nil {
		return response, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return response, err
	}

	llm, err := openai.New(
		openai.WithBaseURL(s.AIURL),
		openai.WithToken(s.AIToken),
		openai.WithEmbeddingModel(s.EmbeddingModel),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		return response, err
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return response, err
	}

	store, err := pgvector.New(
		context.Background(),
		pgvector.WithCollectionName(semanticSearchArgs.Collection),
		pgvector.WithConn(pool),
		pgvector.WithEmbedder(e),
	)
	defer store.Close()
	if err != nil {
		return response, err
	}

	searchResults, err := store.SimilaritySearch(ctx, semanticSearchArgs.Query, s.MaxResults)
	if err != nil {
		return response, err
	}

	for _, result := range searchResults {
		response += fmt.Sprintf("%s\n", result.PageContent)
	}

	return response, nil
}

func init() {
	toolsRegistry = append(toolsRegistry,
		func() (ToolData, error) {
			semanticSearchTool := &SemanticSearchTool{
				AIURL:          os.Getenv("API_URL"),
				AIToken:        os.Getenv("API_TOKEN"),
				DBConnection:   os.Getenv("SEMANTIC_SEARCH_DB_CONNECTION"),
				EmbeddingModel: os.Getenv("SEMANTIC_SEARCH_EMBEDDING_MODEL"),
				MaxResults:     2,
			}

			return ToolData{
				Definition: SemanticSearchDefinition,
				Call:       semanticSearchTool.Call,
			}, nil
		},
	)
}
