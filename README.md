# libagent
A utility library for AI agent development

## Usage
### Configuration
The library is configured through the `Config` structure (see `pkg/config/config.go`)  
Config structure can be initialized using `.env` file with `config.NewConfig()` function call  
See `.envExample`

```go
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("new config")
	}
```

### Agents
There are two agent abstractions in the library for now:  
 - generic (able to use toolsExecutor)
 - simple (just bare LLM)

Agent have two methods:  
```go
func (a *Agent) Run(
	ctx context.Context,
	state []llms.MessageContent,
	opts ...llms.CallOption,
) (llms.MessageContent, error) {...}
```

```go
func (a *Agent) SimpleRun(
	ctx context.Context,
	input string,
	opts ...llms.CallOption,
) (string, error) {...}
```

It can be initialized like this:
```go
	agent := generic.Agent{}

	llm, err := openai.New(
		openai.WithBaseURL(cfg.AIURL),
		openai.WithToken(cfg.AIToken),
		openai.WithModel(cfg.Model),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("new openai api llm")
	}
	agent.LLM = llm
```

### ToolsExecutor
For tools we have a ToolsExecutor abstraction.  
The available tools can be seen in `pkg/tools` directory.  
Each one adds itself to the toolsExecutor registry.  
It can be initialized and fed to the generic agent like this:  
```go
	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg, tools.WithToolsWhitelist(
		tools.ReWOOToolDefinition.Name,
		tools.SemanticSearchDefinition.Name,
		tools.DDGSearchDefinition.Name,
		tools.WebReaderDefinition.Name,
    ...
	))
	if err != nil {
		log.Fatal().Err(err).Msg("new tools executor")
	}
	agent.ToolsExecutor = toolsExecutor
	defer func() {
		if err := toolsExecutor.Cleanup(); err != nil {
			log.Fatal().Err(err).Msg("tools executor cleanup")
		}
	}()
```
Note the deferred `Cleanup` function - for now it used for the shell commands executor tool to clean the temp environment.

The tool can be called directly, not by agent like this:
```go
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
```

### Run
You can SimpleRun (just `string` -> `string`), or Run (`llms.MessageContent` -> `llms.MessageContent`) the agent.  
There are default call options can be configured through the `.env`, which can be used through `config.ConfigToCallOptions(cfg.DefaultCallOptions)...` helper function.  
```go
	result, err := agent.SimpleRun(ctx,
		Prompt, config.ConifgToCallOptions(cfg.DefaultCallOptions)...,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("agent run")
	}
```

### Static analysis
The repository includes the `smbgo` typo-suggestion analyzer. Install and run it with `go vet`:

```bash
go install github.com/skulidropek/GoSuggestMembersAnalyzer/cmd/smbgo@latest
go vet -vettool=$(go env GOPATH)/bin/smbgo ./...
```

The analyzer replaces the standard "not found" diagnostics with detailed messages containing "did you mean" suggestions for selectors, identifiers, and import paths.
