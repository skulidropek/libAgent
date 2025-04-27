## Package: rewoo

Imports:
- "context"
- "libagent/pkg/tools"
- "github.com/JackBekket/langgraphgo/graph/stategraph"
- "github.com/tmc/langchaingo/llms"
- "github.com/tmc/langchaingo/llms/openai"

External data, input sources:
- The code uses an OpenAI LLM for text generation and interaction.

TODOs:
- TODO integrate regular chat flow into ReWOO

### Agent struct
The Agent struct is responsible for managing the interaction between the LLM and the tools executor. It contains the following fields:
- LLM: An instance of the OpenAI LLM.
- ToolsExecutor: An instance of the tools executor.
- graph: A pointer to the runnable graph.

### Run method
The Run method is responsible for running the agent in a given context. It takes a context and a list of messages as input and returns a message and an error.

### SimpleRun method
The SimpleRun method is responsible for running the agent with a simple input. It takes a context and a string as input and returns a string and an error.

### createGraph method
The createGraph method is responsible for creating the graph used by the agent. It returns a pointer to the runnable graph and an error.

### Workflow
The code defines a workflow for the agent, which consists of three nodes: plan, tool, and solve. The plan node is responsible for generating a plan, the tool node is responsible for executing the plan, and the solve node is responsible for solving the task. The workflow also includes conditional edges to handle different situations.

pkg/agent/rewoo/modules.go
Package: rewoo

Imports:
- context
- encoding/json
- fmt
- libagent/pkg/tools
- regexp
- strings
- github.com/tmc/langchaingo/llms

External data, input sources:
- None

TODOs:
- None

Summary:
- The code defines a package called "rewoo" that implements an agent for solving tasks by generating plans and executing tools.
- The agent uses a large language model (LLM) to generate plans and a set of tools to execute the plans.
- The package includes functions for generating plans, executing tools, and solving tasks.
- The agent follows a specific workflow: first, it generates a plan, then it executes the tools in the plan, and finally, it solves the task using the results from the tool executions.
- The package also includes functions for routing the agent between different stages of the workflow.
- The code uses regular expressions to extract information from the LLM's output and stores the results in a map.
- The agent uses a context to manage the execution of the tools and the LLM.
- The package includes functions for handling errors and exceptions.
- The code is well-documented and includes comments explaining the purpose of each function and variable.
- The package is designed to be modular and extensible, allowing users to add new tools and functionalities.

pkg/agent/rewoo/agent.go
- agent.go
- modules.go
