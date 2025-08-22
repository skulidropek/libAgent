package main

import (
	"fmt"

	"github.com/Swarmind/libagent/examples/codemonkey/pkg/reviewer"
	"github.com/Swarmind/libagent/pkg/util"
)

func main() {

	//githellper goroutine for incoming issues

	task := reviewer.GatherInfo("Save only groupchats to the database", "Hellper")
	task = util.RemoveThinkTag(task)
	fmt.Println("Output: ", task)

	//planner CreatePlan()

	//executor Execute()

	//if Execute() ==err CreatePlan(err text) x 2

}
