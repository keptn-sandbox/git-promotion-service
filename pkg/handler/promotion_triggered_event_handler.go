package handler

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"k8s.io/client-go/kubernetes"
)

const PromotionTaskName = "promotion"
const githubPathRegexp = "^/[a-zA-Z0-9-]+/[a-zA-Z-_.]+$"

type PromotionTriggeredEventHandler struct {
	keptn *keptnv2.Keptn
}

type PromotionTriggeredEventData struct {
	keptnv2.EventData
	Promotion Promotion `json:"promotion"`
}

type Promotion struct {
	ToStage string `json:"tostage"`
	Repository string `json:"repository"`
	SecretName string `json:"secretname"`
	Strategy string `json:"strategy"`
}

// NewPromotionTriggeredEventHandler returns a new promotion.triggered event handler
func NewPromotionTriggeredEventHandler(keptn *keptnv2.Keptn) *PromotionTriggeredEventHandler {
	return &PromotionTriggeredEventHandler{keptn: keptn}
}

// IsTypeHandled godoc
func (a *PromotionTriggeredEventHandler) IsTypeHandled(event cloudevents.Event) bool {
	return event.Type() == keptnv2.GetTriggeredEventType(PromotionTaskName)
}

// Handle godoc
func (a *PromotionTriggeredEventHandler) Handle(event cloudevents.Event, keptnHandler *keptnv2.Keptn) {
	data := &PromotionTriggeredEventData{}
	if err := event.DataAs(data); err != nil {
		logger.WithError(err).Error("failed to parse PromotionTriggeredEventData")
		return
	}
	outgoingEvents := a.handlePromotionTriggeredEvent(*data, event.Context.GetID(), keptnHandler.KeptnContext)
	sendEvents(keptnHandler, outgoingEvents)
}

func (a *PromotionTriggeredEventHandler) handlePromotionTriggeredEvent(inputEvent PromotionTriggeredEventData,
	triggeredID, shkeptncontext string) []cloudevents.Event {
	outgoingEvents := make([]cloudevents.Event, 0)

	startedEvent := a.getPromotionStartedEvent(inputEvent, triggeredID, shkeptncontext)
	outgoingEvents = append(outgoingEvents, *startedEvent)
	logger.Printf("Start promoting from %s to %s in repository %s with strategy %s. The accesstoken should be found in secret %s", inputEvent.Stage, inputEvent.Promotion.ToStage, inputEvent.Promotion.Strategy, inputEvent.Promotion.Repository, inputEvent.Promotion.SecretName)
	var status keptnv2.StatusType
	var result keptnv2.ResultType
	var message string
	if val := validateInputEvent(inputEvent) ; len(val) > 0 {
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "validation error: " + strings.Join(val, ",")
	} else if accessToken, err := getAccessToken(inputEvent.Promotion.SecretName) ; err != nil {
		logger.Printf("error while reading secret with name %s: %s", inputEvent.Promotion.SecretName, err)
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while reading secret"
	} else if msg,err := openPullRequest(inputEvent.Promotion.Repository, inputEvent.Stage, inputEvent.Promotion.ToStage, accessToken) ; err != nil {
		logger.Printf("could not open pull request on repository %s: %s", inputEvent.Promotion.Repository, err)
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while opening pull request"
	} else {
		status = keptnv2.StatusSucceeded
		result = keptnv2.ResultPass
		message = msg
	}
    finishedEvent := a.getPromotionFinishedEvent(inputEvent, status, result, message, triggeredID, shkeptncontext)
	outgoingEvents = append(outgoingEvents, *finishedEvent)

	return outgoingEvents
}

func openPullRequest(repositoryUrl, fromBranch, toBranch, accessToken string) (message string, err error) {
	owner, repo, err := getGithubOwnerRepository(repositoryUrl)
	if err != nil {
		return message, err
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	compare, _, err := client.Repositories.CompareCommits(ctx, owner, repo, toBranch, fromBranch)
	if err != nil {
		return message, err
	}
	if len(compare.Commits) == 0 {
		return fmt.Sprintf("no difference between branches %s and %s found => nothing todo", fromBranch, toBranch), nil
	}
	pull, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head:        fromBranch,
		Base:        toBranch,
	})
	if err != nil {
		return message, err
	}
	if len(pull) > 0 {
		return fmt.Sprintf("pull request already open: %s", *pull[0].HTMLURL), nil
	}
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String(fmt.Sprintf("Promote to stage %s", toBranch)),
		Head:                &fromBranch,
		Base:                &toBranch,
		Body:                github.String("> Pull Request opened by keptn "),
	})
	if err != nil {
		return message, err
	}
	return fmt.Sprintf("opened pull request %s", *pr.HTMLURL), nil
}

func getGithubOwnerRepository(raw string) (owner,repository string, err error) {
	u,err := url.Parse(raw)
	if err != nil {
		return owner, repository, err
	}
	splittedUrl := strings.Split(u.Path, "/")
	return splittedUrl[1], splittedUrl[2], nil
}

func (a *PromotionTriggeredEventHandler) getPromotionStartedEvent(inputEvent PromotionTriggeredEventData, triggeredID, shkeptncontext string) *cloudevents.Event {
	promotionStartedEvent := keptnv2.EventData{
			Project: inputEvent.Project,
			Stage:   inputEvent.Stage,
			Service: inputEvent.Service,
			Labels:  inputEvent.Labels,
			Status:  keptnv2.StatusSucceeded,
			Message: "Promotion started",
	}
	return getCloudEvent(promotionStartedEvent, keptnv2.GetStartedEventType(PromotionTaskName), shkeptncontext, triggeredID)
}

func (a *PromotionTriggeredEventHandler) getPromotionFinishedEvent(inputEvent PromotionTriggeredEventData,
	status keptnv2.StatusType, result keptnv2.ResultType, message string, triggeredID, shkeptncontext string) *cloudevents.Event {
	promotionFinishedEvent := keptnv2.EventData{
			Project: inputEvent.Project,
			Stage:   inputEvent.Stage,
			Service: inputEvent.Service,
			Labels:  inputEvent.Labels,
			Status:  status,
			Result:  result,
			Message: message,
	}
	return getCloudEvent(promotionFinishedEvent, keptnv2.GetFinishedEventType(PromotionTaskName), shkeptncontext, triggeredID)
}

func validateInputEvent(inputEvent PromotionTriggeredEventData) (validationErrrors []string) {
	if inputEvent.Promotion.ToStage == "" {
		validationErrrors = append(validationErrrors, `"tostage" missing`)
	}
	if inputEvent.Promotion.Strategy == "" {
		validationErrrors = append(validationErrrors, `"strategy" missing`)
	} else if inputEvent.Promotion.Strategy != "branches" {
		validationErrrors = append(validationErrrors, `"strategy" invalid`)
	}
	if inputEvent.Promotion.SecretName == "" {
		validationErrrors = append(validationErrrors, `"secretname" missing`)
	}
	if inputEvent.Promotion.Repository == "" {
		validationErrrors = append(validationErrrors, `"repository" missing`)
	} else {
		u, err := url.Parse(inputEvent.Promotion.Repository)
		if err != nil {
			validationErrrors = append(validationErrrors, `"repository" is not a valid URL`)
		} else {
			if u.Scheme != "https" || u.Host != "github.com"  {
				validationErrrors = append(validationErrrors, `"repository" must be a "https" url to a repository on github.com`)
			} else if matched, err := regexp.MatchString(githubPathRegexp, u.Path) ; err!=nil || !matched {
				validationErrrors = append(validationErrrors, `"repository" must be a "https" url to a repository on github.com`)
			}
		}
	}
	return validationErrrors
}

func getAccessToken(secretName string) (accessToken string, err error) {
	if client, err := createKubeAPI() ; err != nil {
		return accessToken, err
	} else if secret, err := client.CoreV1().Secrets(os.Getenv("K8S_NAMESPACE")).Get(context.Background(), secretName, v1.GetOptions{}) ; err != nil {
		return accessToken, err
	} else {
		return string(secret.Data["access-token"]), nil
	}
}

func createKubeAPI() (*kubernetes.Clientset, error) {
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeAPI, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubeAPI, nil
}
