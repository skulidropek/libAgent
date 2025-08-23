package main

import (
	"fmt"

	githubservice "github.com/Swarmind/libagent/examples/codemonkey/pkg/github"
	"github.com/Swarmind/libagent/examples/codemonkey/pkg/reviewer"
	utility "github.com/Swarmind/libagent/examples/codemonkey/pkg/util"
	"github.com/Swarmind/libagent/pkg/util"
)

func main() {

	es := &githubservice.EventsService{
		GithubAPI: githubservice.ConstructGithubApi(),
		Ichan:     make(chan githubservice.IssueEvent, 10),
	}

	go es.StartWebhookHandler(utility.GetEnv("LISTEN_ADDR"))

	for issue := range es.Ichan {
		fmt.Printf("Got issue: %s\n", issue.RepoName)

		task := reviewer.GatherInfo(issue.IssueText, issue.RepoName)
		task = util.RemoveThinkTag(task)

		fmt.Println("Reviewer result: ", task)

	}
}
