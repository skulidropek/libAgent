package main

import (
	"fmt"

	"github.com/Swarmind/libagent/examples/codemonkey/pkg/planner"
	"github.com/Swarmind/libagent/examples/codemonkey/pkg/reviewer"
	"github.com/Swarmind/libagent/pkg/util"
)

func main() {

	//issue flow

	/* es := &githubservice.EventsService{
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
	*/

	//test stuff
	task := reviewer.GatherInfo("Change hello message to Can I haz cheeseburger?", "Hellper")
	task = util.RemoveThinkTag(task)
	plan := planner.Plan(task)
	fmt.Println("Planner result: ", plan)
}
