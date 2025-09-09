package main

import (
	"fmt"

	"github.com/Swarmind/libagent/examples/codemonkey/pkg/executor"
	"github.com/Swarmind/libagent/examples/codemonkey/pkg/planner"
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
	/* task := reviewer.GatherInfo("Change hello message to Can I haz cheeseburger?", "Hellper")
	task = util.RemoveThinkTag(task)
	plan := planner.PlanGitHelper(task)
	fmt.Println("Planner result: ", plan)
	*/
	plan := planner.PlanCLIExecutor(`Find your current directory, then list all filenames of it. If it contains .txt files, delete them.`)
	fmt.Println(plan)
	executor.ExecuteCommands(plan)
}
