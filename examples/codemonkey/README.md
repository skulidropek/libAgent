# Code Monkey Agent Example

This example demonstrates how to use the Code Monkey Agent to automatically resolve issues in a codebase.

## Prerequisites

1. Go 1.21 or later
2. PostgreSQL database (for semantic search)
3. OpenAI API key
4. DuckDuckGo API access (optional, for web search)

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL:
   - Install PostgreSQL if you haven't already
   - Create a database named `libagent`
   - Update the `SEMANTIC_SEARCH_DB_CONNECTION` in `.env` if needed

3. Configure environment variables:
   - Copy the `.env` file to the project root
   - Update the following variables:
     - `AI_TOKEN`: Your OpenAI API key
     - `SEMANTIC_SEARCH_DB_CONNECTION`: Your PostgreSQL connection string
     - Other variables as needed

## Running the Example

1. Build the example:
```bash
go build -o codemonkey examples/codemonkey/main.go
```

2. Run the example:
```bash
./codemonkey
```

The example will:
1. Initialize the Code Monkey Agent
2. Process a sample issue about an authentication system bug
3. Use semantic search to find relevant code
4. Generate a solution plan
5. Execute the plan using available tools
6. Report the results

## Customizing the Example

To test with your own issues:

1. Modify the `issue` variable in `main.go` with your problem description
2. Run the example again

Example issue format:
```go
issue := `Your detailed issue description here.
Include any error messages, stack traces, or relevant context.`
```

## Available Tools

The Code Monkey Agent uses several tools:
- Semantic Search: For finding relevant code
- ReWOO: For complex reasoning and planning
- Command Executor: For file and system operations
- Web Search: For additional research
- Web Reader: For reading documentation

## Troubleshooting

1. If you get database connection errors:
   - Verify PostgreSQL is running
   - Check the connection string in `.env`
   - Ensure the database exists

2. If you get API errors:
   - Verify your API keys are correct
   - Check your API rate limits
   - Ensure you have internet connectivity

3. If tools are not working:
   - Check the tool-specific configuration in `.env`
   - Verify the tool is not disabled
   - Check the logs for specific error messages 