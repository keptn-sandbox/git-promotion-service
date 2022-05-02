package handler

import (
	"context"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	api "github.com/keptn/go-utils/pkg/api/utils"
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

const GitPromotionTaskName = "git-promotion"
const githubPathRegexp = "^/[a-zA-Z0-9-]+/[a-zA-Z-_.]+$"

type GitPromotionTriggeredEventHandler struct {
	keptn *keptnv2.Keptn
}

type GitPromotionTriggeredEventData struct {
	keptnv2.EventData
	GitPromotion GitPromotion `json:"git-promotion"`
}

type GitPromotion struct {
	Repository string `json:"repository"`
	SecretName string `json:"secretname"`
	Strategy   string `json:"strategy"`
}

// NewGitPromotionTriggeredEventHandler returns a new promotion.triggered event handler
func NewGitPromotionTriggeredEventHandler(keptn *keptnv2.Keptn) *GitPromotionTriggeredEventHandler {
	return &GitPromotionTriggeredEventHandler{keptn: keptn}
}

// IsTypeHandled godoc
func (a *GitPromotionTriggeredEventHandler) IsTypeHandled(event cloudevents.Event) bool {
	return event.Type() == keptnv2.GetTriggeredEventType(GitPromotionTaskName)
}

// Handle godoc
func (a *GitPromotionTriggeredEventHandler) Handle(event cloudevents.Event, keptnHandler *keptnv2.Keptn) {
	data := &GitPromotionTriggeredEventData{}
	if err := event.DataAs(data); err != nil {
		logger.WithError(err).Error("failed to parse GitPromotionTriggeredEventData")
		return
	}
	outgoingEvents := a.handleGitPromotionTriggeredEvent(*data, event.Context.GetID(), keptnHandler.KeptnContext)
	sendEvents(keptnHandler, outgoingEvents)
}

func (a *GitPromotionTriggeredEventHandler) handleGitPromotionTriggeredEvent(inputEvent GitPromotionTriggeredEventData,
	triggeredID, shkeptncontext string) []cloudevents.Event {
	outgoingEvents := make([]cloudevents.Event, 0)

	startedEvent := a.getGitPromotionStartedEvent(inputEvent, triggeredID, shkeptncontext)
	outgoingEvents = append(outgoingEvents, *startedEvent)
	logger.WithField("func", "handleGitPromotionTriggeredEvent").Infof("start promoting from %s in repository %s with strategy %s. The accesstoken should be found in secret %s", inputEvent.Stage, inputEvent.GitPromotion.Repository, inputEvent.GitPromotion.Strategy, inputEvent.GitPromotion.SecretName)
	var status keptnv2.StatusType
	var result keptnv2.ResultType
	var message string
	if val := validateInputEvent(inputEvent); len(val) > 0 {
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "validation error: " + strings.Join(val, ",")
	} else if accessToken, err := getAccessToken(inputEvent.GitPromotion.SecretName); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Errorf("handleGitPromotionTriggeredEvent: error while reading secret with name %s", inputEvent.GitPromotion.SecretName)
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while reading secret"
	} else if nextStage, err := getNextStage(inputEvent.Project, inputEvent.Stage); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Error("handleGitPromotionTriggeredEvent: error while reading nextStage")
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while reading nextStage"
	} else if msg, err := openPullRequest(inputEvent.GitPromotion.Repository, inputEvent.Stage, nextStage, accessToken); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Errorf("handleGitPromotionTriggeredEvent: could not open pull request on repository %s", inputEvent.GitPromotion.Repository)
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while opening pull request"
	} else {
		status = keptnv2.StatusSucceeded
		result = keptnv2.ResultPass
		message = msg
	}
	finishedEvent := a.getGitPromotionFinishedEvent(inputEvent, status, result, message, triggeredID, shkeptncontext)
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
		logger.WithField("func", "openPullRequest").Infof("no difference found in repo %s from branch %s to %s", repositoryUrl, fromBranch, toBranch)
		return fmt.Sprintf("no difference between branches %s and %s found => nothing todo", fromBranch, toBranch), nil
	}
	pull, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head: fromBranch,
		Base: toBranch,
	})
	if err != nil {
		return message, err
	}
	if len(pull) > 0 {
		logger.WithField("func", "openPullRequest").Infof("pull request in repo %s from branch %s to %s already open with id %d", repositoryUrl, fromBranch, toBranch, *pull[0].Number)
		return fmt.Sprintf("pull request already open: %s", *pull[0].HTMLURL), nil
	}
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Promote to stage %s", toBranch)),
		Head:  &fromBranch,
		Base:  &toBranch,
		Body:  github.String("> Pull Request opened by keptn "),
	})
	if err != nil {
		return message, err
	}
	logger.WithField("func", "openPullRequest").Infof("opened pull request %d in repo %s from branch %s to %s", *pr.Number, repositoryUrl, fromBranch, toBranch)
	return fmt.Sprintf("opened pull request %s", *pr.HTMLURL), nil
}

func getGithubOwnerRepository(raw string) (owner, repository string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return owner, repository, err
	}
	splittedUrl := strings.Split(u.Path, "/")
	return splittedUrl[1], splittedUrl[2], nil
}

func (a *GitPromotionTriggeredEventHandler) getGitPromotionStartedEvent(inputEvent GitPromotionTriggeredEventData, triggeredID, shkeptncontext string) *cloudevents.Event {
	promotionStartedEvent := keptnv2.EventData{
		Project: inputEvent.Project,
		Stage:   inputEvent.Stage,
		Service: inputEvent.Service,
		Labels:  inputEvent.Labels,
		Status:  keptnv2.StatusSucceeded,
		Message: "GitPromotion started",
	}
	return getCloudEvent(promotionStartedEvent, keptnv2.GetStartedEventType(GitPromotionTaskName), shkeptncontext, triggeredID)
}

func (a *GitPromotionTriggeredEventHandler) getGitPromotionFinishedEvent(inputEvent GitPromotionTriggeredEventData,
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
	return getCloudEvent(promotionFinishedEvent, keptnv2.GetFinishedEventType(GitPromotionTaskName), shkeptncontext, triggeredID)
}

func validateInputEvent(inputEvent GitPromotionTriggeredEventData) (validationErrrors []string) {
	if inputEvent.GitPromotion.Strategy == "" {
		validationErrrors = append(validationErrrors, `"strategy" missing`)
	} else if inputEvent.GitPromotion.Strategy != "branches" {
		validationErrrors = append(validationErrrors, `"strategy" invalid`)
	}
	if inputEvent.GitPromotion.SecretName == "" {
		validationErrrors = append(validationErrrors, `"secretname" missing`)
	}
	if inputEvent.GitPromotion.Repository == "" {
		validationErrrors = append(validationErrrors, `"repository" missing`)
	} else {
		u, err := url.Parse(inputEvent.GitPromotion.Repository)
		if err != nil {
			validationErrrors = append(validationErrrors, `"repository" is not a valid URL`)
		} else {
			if u.Scheme != "https" || u.Host != "github.com" {
				validationErrrors = append(validationErrrors, `"repository" must be a "https" url to a repository on github.com`)
			} else if matched, err := regexp.MatchString(githubPathRegexp, u.Path); err != nil || !matched {
				validationErrrors = append(validationErrrors, `"repository" must be a "https" url to a repository on github.com`)
			}
		}
	}
	logger.WithField("func", "validateInputEvent").Infof("validation for %s/%s/%s finished with %d validation errors", inputEvent.Project, inputEvent.Stage, inputEvent.Service, len(validationErrrors))
	return validationErrrors
}

func getAccessToken(secretName string) (accessToken string, err error) {
	if client, err := createKubeAPI(); err != nil {
		return accessToken, err
	} else if secret, err := client.CoreV1().Secrets(os.Getenv("K8S_NAMESPACE")).Get(context.Background(), secretName, v1.GetOptions{}); err != nil {
		return accessToken, err
	} else {
		logger.WithField("func", "getAccessToken").Infof("found access-token with length %d in secret %s", len(secret.Data["access-token"]), secret.Name)
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

func getNextStage(project string, stage string) (nextStage string, err error) {
	apiSet, err := api.New(os.Getenv("API_BASE_URL"), api.WithAuthToken(os.Getenv("API_AUTH_TOKEN")))
	if err != nil {
		logger.WithField("func", "getNextStage").WithError(err).Errorf("could not get apiSet for project %s with stage %s", project, stage)
		return nextStage, err
	}
	stages, err := apiSet.StagesV1().GetAllStages(project)
	if err != nil {
		logger.WithField("func", "getNextStage").WithError(err).Errorf("could not get all stages for project %s with stage %s", project, stage)
		return nextStage, err
	}
	for i, s := range stages {
		if s.StageName == stage {
			if len(stages) <= (i + 1) {
				err = errors.New(fmt.Sprintf("no stage defined after stage %s", stage))
				logger.WithField("func", "getNextStage").WithError(err).Errorf("no next stage found for project %s with stage %s", project, stage)
				return nextStage, err
			}
			logger.WithField("func", "getNextStage").Infof("next stage %s found for project %s and stage %s", stages[i+1].StageName, project, stage)
			return stages[i+1].StageName, nil
		}
	}
	err = errors.New(fmt.Sprintf("stage %s not found", stage))
	logger.WithField("func", "getNextStage").WithError(err).Errorf("stage %s not found for project %s", stage, project)
	return nextStage, err
}
