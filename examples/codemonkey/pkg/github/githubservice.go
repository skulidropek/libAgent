package githubservice

import (
	"context"
	"net/http"
	"os"
	"strconv"

	utility "github.com/Swarmind/libagent/examples/codemonkey/pkg/util"

	"github.com/cbrgm/githubevents/v2/githubevents"
	"github.com/google/go-github/v74/github"
	"github.com/rs/zerolog/log"
)

type GithubAPI struct {
	AppId  int64
	PkPath string
}

type EventsService struct {
	GithubAPI GithubAPI
	Ichan     chan IssueEvent
}

type IssueEvent struct {
	RepoName  string
	IssueText string
}

func (es *EventsService) StartWebhookHandler(port string) {
	handle := githubevents.New(os.Getenv("WEBHOOK_SECRET_KEY"))
	handle.OnIssuesEventAny(es.IssuesEventAnyHandler)
	http.HandleFunc(utility.GetEnv("WEBHOOK_ROUTE"), func(w http.ResponseWriter, r *http.Request) {
		if err := handle.HandleEventRequest(r); err != nil {
			log.Warn().Err(err).Msg("handle github event request")
		}
	})

	if err := http.ListenAndServe(utility.GetEnv("LISTEN_ADDR"), nil); err != nil {
		log.Fatal().Err(err).Msg("listen and serve")

	}
}

func ConstructGithubApi() GithubAPI {
	appIdStr := utility.GetEnv("APP_ID")
	appId, err := strconv.ParseInt(appIdStr, 10, 64)
	if err != nil {
		log.Fatal().Err(err).Msg("parsing APP_ID env variable")
	}

	pkPath := utility.GetEnv("PRIVKEY_PATH")
	if _, err := os.Stat(pkPath); err != nil {
		log.Fatal().Err(err).Msg("could not read file for PRIVKEY_PATH env variable")
	}

	return GithubAPI{
		AppId:  appId,
		PkPath: pkPath,
	}
}

func (es *EventsService) IssuesEventAnyHandler(ctx context.Context,
	_, _ string,
	event *github.IssuesEvent) error {
	issue := event.GetIssue()
	issueEvent := IssueEvent{
		RepoName:  event.GetRepo().GetFullName(),
		IssueText: issue.GetTitle() + "\n" + issue.GetBody(),
	}

	// Отправляем событие в канал
	es.Ichan <- issueEvent
	return nil
}

func GetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Msgf("environment key %s is empty", key)
	}
	return val
}
