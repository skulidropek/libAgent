# tools

This package provides tools for semantic search and DuckDuckGo search.

pkg/tools/ddgSearch.go
pkg/tools/semanticSearch.go
pkg/tools/tools.go

The `tools` package offers a set of tools for semantic search and DuckDuckGo search. The `semanticSearch` tool performs semantic search in a vector store of saved code blobs, while the `DDGSearch` tool wraps around DuckDuckGo Search. The `ToolsExecutor` struct manages and executes these tools, allowing users to easily access and utilize their functionalities.

The `semanticSearch` tool leverages an OpenAI API for the LLM and a PostgreSQL database for the vector store. It takes a query and a collection name as input and returns matching file contents. The tool uses an OpenAI LLM to generate embeddings for the query and the documents in the vector store, and then performs the similarity search using a vector store implementation from the langchaingo library.

The `DDGSearch` tool wraps around DuckDuckGo Search, taking a search query as input and returning the search results.

The `ToolsExecutor` struct is responsible for managing and executing these tools. It stores a map of tools, where each tool is associated with its definition and a call function. The `Execute` function takes a context and a ToolCall object as input and returns a ToolCallResponse object. It first checks if the tool exists in the Tools map. If it does, it calls the tool's call function with the given arguments and returns the response. If the tool does not exist, it returns an error.

The `ToolsList` function returns a list of all the tools available in the ToolsExecutor, while the `ToolsPromptDesc` function returns a string describing all the tools available in the ToolsExecutor.

The package provides a convenient way to perform semantic search and DuckDuckGo search, making it easy for users to find the information they need.

