package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Swarmind/libagent/pkg/agent/codemonkey"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeMonkeyAgent(t *testing.T) {
	// Create a test repository
	testRepoPath, err := createTestRepo(t)
	require.NoError(t, err)
	defer os.RemoveAll(testRepoPath)

	// Initialize config
	cfg := config.Config{
		AIURL:   os.Getenv("AI_URL"),
		AIToken: os.Getenv("AI_TOKEN"),
		Model:   os.Getenv("AI_MODEL"),
	}

	// Create agent
	ctx := context.Background()
	agent, err := codemonkey.NewCodeMonkeyAgent(ctx, cfg)
	require.NoError(t, err)
	defer agent.Cleanup()

	// Test cases
	tests := []struct {
		name        string
		repoURL     string
		issueText   string
		expectError bool
	}{
		{
			name:    "Simple bug fix",
			repoURL: testRepoPath,
			issueText: `Fix authentication bug
The authentication system is not working properly. Users are unable to log in with their credentials.
The error occurs in the login handler and seems to be related to password validation.`,
			expectError: false,
		},
		{
			name:    "Invalid repository",
			repoURL: "https://github.com/nonexistent/repo.git",
			issueText: `Fix a bug
This is a test issue.`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agent.ResolveIssue(ctx, tt.repoURL, tt.issueText)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Changes have been pushed to branch")
		})
	}
}

// createTestRepo creates a test repository with some initial code
func createTestRepo(t *testing.T) (string, error) {
	// Create a temporary directory for the test repository
	repoPath, err := os.MkdirTemp("", "codemonkey_test_*")
	if err != nil {
		return "", err
	}

	// Initialize git repository
	commands := []string{
		"git init",
		"git config user.name 'Test User'",
		"git config user.email 'test@example.com'",
	}

	for _, cmd := range commands {
		if err := runCommand(repoPath, cmd); err != nil {
			os.RemoveAll(repoPath)
			return "", err
		}
	}

	// Create a simple Go file with a bug
	mainFile := filepath.Join(repoPath, "main.go")
	content := `package main

import "fmt"

func main() {
	// Bug: Authentication always fails
	if authenticate("user", "pass") {
		fmt.Println("Login successful")
	} else {
		fmt.Println("Login failed")
	}
}

func authenticate(username, password string) bool {
	// Bug: Always returns false
	return false
}
`
	if err := os.WriteFile(mainFile, []byte(content), 0644); err != nil {
		os.RemoveAll(repoPath)
		return "", err
	}

	// Create initial commit
	commands = []string{
		"git add .",
		"git commit -m 'Initial commit'",
	}

	for _, cmd := range commands {
		if err := runCommand(repoPath, cmd); err != nil {
			os.RemoveAll(repoPath)
			return "", err
		}
	}

	return repoPath, nil
}

// runCommand executes a shell command in the specified directory
func runCommand(dir, cmd string) error {
	execCmd := exec.Command("bash", "-c", cmd)
	execCmd.Dir = dir
	return execCmd.Run()
} 