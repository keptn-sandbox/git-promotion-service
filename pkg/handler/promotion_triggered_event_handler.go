package handler

import (
	"context"
	"errors"
	"fmt"
	promotionconfig "keptn/git-promotion-service/pkg/config"
	"keptn/git-promotion-service/pkg/model"
	"keptn/git-promotion-service/pkg/promoter"
	"keptn/git-promotion-service/pkg/replacer"
	"keptn/git-promotion-service/pkg/repoaccess"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const GitPromotionTaskName = "git-promotion"
const keptnPullRequestTitlePrefix = "keptn:"
const configurationResource = GitPromotionTaskName + ".yaml"

type GitPromotionTriggeredEventHandler struct {
	keptn      *keptnv2.Keptn
	api        *api.APISet
	kubeClient *kubernetes.Clientset
}

type GitPromotionTriggeredEventData struct {
	keptnv2.EventData
}

// NewGitPromotionTriggeredEventHandler returns a new GitPromotionTriggeredEventHandler
func NewGitPromotionTriggeredEventHandler(keptn *keptnv2.Keptn, api *api.APISet, kubeClient *kubernetes.Clientset) *GitPromotionTriggeredEventHandler {
	return &GitPromotionTriggeredEventHandler{keptn: keptn, api: api, kubeClient: kubeClient}
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
	outgoingEvents := a.handleGitPromotionTriggeredEvent(event, *data, event.Context.GetID(), keptnHandler.KeptnContext)
	sendEvents(keptnHandler, outgoingEvents)
}

func (a *GitPromotionTriggeredEventHandler) handleGitPromotionTriggeredEvent(event cloudevents.Event, inputEvent GitPromotionTriggeredEventData,
	triggeredID, shkeptncontext string) []cloudevents.Event {
	logger.WithField("func", "handleGitPromotionTriggeredEvent").Infof("start promoting service %s in project %s from stage %s", inputEvent.Service, inputEvent.Stage, inputEvent.Project)
	if err := a.keptn.SendCloudEvent(*a.getGitPromotionStartedEvent(inputEvent, triggeredID, shkeptncontext)); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Errorf("sending started event failed")
		return []cloudevents.Event{*a.getGitPromotionFinishedEvent(inputEvent, keptnv2.StatusErrored, keptnv2.ResultFailed, "sending starting event failed", triggeredID, shkeptncontext, nil)}
	}
	outgoingEvents := make([]cloudevents.Event, 0)
	var nextStage string
	if nextStageTemp, err := a.getNextStage(inputEvent.Project, inputEvent.Stage); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Error("handleGitPromotionTriggeredEvent: error while reading nextStage")
		return []cloudevents.Event{*a.getGitPromotionFinishedEvent(inputEvent, keptnv2.StatusErrored, keptnv2.ResultFailed, "error while reading nextStage", triggeredID, shkeptncontext, nil)}
	} else {
		nextStage = nextStageTemp
	}
	config := a.getMergedConfiguration(inputEvent.GetProject(), inputEvent.GetStage(), nextStage, inputEvent.GetService())
	logger.WithField("func", "handleGitPromotionTriggeredEvent").Infof("using git promotion config: strategy: %s, repository: %s, secret: %s", toString(config.Spec.Strategy), toString(config.Spec.Target.Repo), toString(config.Spec.Target.Secret))
	var status keptnv2.StatusType
	var result keptnv2.ResultType
	var message string
	var prLink *string
	if vs := promotionconfig.NewValidator().Validate(config); len(vs) > 0 {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").Errorf("validation of configuration failed: %s", strings.Join(vs, ","))
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "validation error: " + strings.Join(vs, ",")
	} else if accessToken, err := a.getAccessToken(*config.Spec.Target.Secret); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Errorf("handleGitPromotionTriggeredEvent: error while reading secret with name %s", *config.Spec.Target.Secret)
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while reading secret"
	} else if client, err := repoaccess.NewClient(accessToken, *config.Spec.Target.Repo); err != nil {
		logger.WithField("func", "handleGitPromotionTriggeredEvent").WithError(err).Errorf("handleGitPromotionTriggeredEvent: error while creating client for repo")
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "error while reading secret"
	} else if *config.Spec.Strategy == model.StrategyBranch {
		status, result, message, prLink = handleBranchStrategy(client, inputEvent, config, shkeptncontext, nextStage)
	} else if *config.Spec.Strategy == model.StrategyFlatPR {
		status, result, message, prLink = handleFlatPRStrategy(client, event, inputEvent, config, shkeptncontext, nextStage)
	} else {
		status = keptnv2.StatusErrored
		result = keptnv2.ResultFailed
		message = "unimplemented strategy"
	}
	finishedEvent := a.getGitPromotionFinishedEvent(inputEvent, status, result, message, triggeredID, shkeptncontext, prLink)
	outgoingEvents = append(outgoingEvents, *finishedEvent)
	return outgoingEvents
}

func handleFlatPRStrategy(client repoaccess.Client, event cloudevents.Event, inputEvent GitPromotionTriggeredEventData, config model.PromotionConfig, shkeptncontext, nextStage string) (status keptnv2.StatusType, result keptnv2.ResultType, message string, prLink *string) {
	p := promoter.NewFlatPrPromoter(client)
	if msg, prlink, err := p.Promote(*config.Spec.Target.Repo, replacer.ConvertToMap(event), "main",
		buildBranchName(inputEvent.Stage, nextStage, shkeptncontext),
		buildTitle(shkeptncontext, nextStage),
		buildBody(shkeptncontext, inputEvent.Project, inputEvent.Service, inputEvent.Stage), config.Spec.Paths); err != nil {
		logger.WithField("func", "handleFlatPRStrategy").WithError(err).Errorf("flat pr strategy failed on repository %s", *config.Spec.Target.Repo)
		return keptnv2.StatusErrored, keptnv2.ResultFailed, "error while opening pull request", nil
	} else {
		return keptnv2.StatusSucceeded, keptnv2.ResultPass, msg, prlink
	}
}

func handleBranchStrategy(client repoaccess.Client, inputEvent GitPromotionTriggeredEventData, config model.PromotionConfig, shkeptncontext, nextStage string) (status keptnv2.StatusType, result keptnv2.ResultType, message string, prLink *string) {
	p := promoter.NewBranchPromoter(client, keptnPullRequestTitlePrefix)
	if msg, prLink, err := p.Promote(*config.Spec.Target.Repo, inputEvent.Stage, nextStage, buildTitle(shkeptncontext, nextStage), buildBody(shkeptncontext, inputEvent.Project, inputEvent.Service, inputEvent.Stage)); err != nil {
		logger.WithField("func", "handleBranchStrategy").WithError(err).Errorf("branch strategy failed on repository %s", *config.Spec.Target.Repo)
		return keptnv2.StatusErrored, keptnv2.ResultFailed, "error while opening pull request", nil
	} else {
		return keptnv2.StatusSucceeded, keptnv2.ResultPass, msg, prLink
	}
}

func buildTitle(keptncontext, nextStage string) string {
	return fmt.Sprintf("%s Promote to stage %s (ctx: %s)", keptnPullRequestTitlePrefix, nextStage, keptncontext)
}

func buildBody(keptncontext, projectName, serviceName, stage string) string {
	return fmt.Sprintf(`Opened by cloud-automation sequence [%s](%s/bridge/project/%s/sequence/%s/stage/%s).

Project: *%s* 
Service: *%s* 
Stage: *%s*`, keptncontext, os.Getenv("EXTERNAL_URL"), projectName, keptncontext, stage, projectName, serviceName, stage)
}

func buildBranchName(stage string, nextStage string, shkeptncontext string) string {
	return fmt.Sprintf("promote/%s_%s-%s", stage, nextStage, shkeptncontext)
}

func (a *GitPromotionTriggeredEventHandler) getGitPromotionStartedEvent(inputEvent GitPromotionTriggeredEventData, triggeredID, shkeptncontext string) *cloudevents.Event {
	gitPromotionStartedEvent := keptnv2.EventData{
		Project: inputEvent.Project,
		Stage:   inputEvent.Stage,
		Service: inputEvent.Service,
		Labels:  inputEvent.Labels,
		Status:  keptnv2.StatusSucceeded,
		Message: "GitPromotion started",
	}
	return getCloudEvent(gitPromotionStartedEvent, keptnv2.GetStartedEventType(GitPromotionTaskName), shkeptncontext, triggeredID)
}

func (a *GitPromotionTriggeredEventHandler) getGitPromotionFinishedEvent(inputEvent GitPromotionTriggeredEventData,
	status keptnv2.StatusType, result keptnv2.ResultType, message string, triggeredID, shkeptncontext string, prLink *string) *cloudevents.Event {
	labels := inputEvent.Labels
	if prLink != nil {
		labels["pullrequest"] = *prLink
	}
	gitPromotionFinishedEvent := keptnv2.EventData{
		Project: inputEvent.Project,
		Stage:   inputEvent.Stage,
		Service: inputEvent.Service,
		Labels:  labels,
		Status:  status,
		Result:  result,
		Message: message,
	}
	return getCloudEvent(gitPromotionFinishedEvent, keptnv2.GetFinishedEventType(GitPromotionTaskName), shkeptncontext, triggeredID)
}

func (a *GitPromotionTriggeredEventHandler) getAccessToken(secretName string) (accessToken string, err error) {
	if secret, err := a.kubeClient.CoreV1().Secrets(os.Getenv("K8S_NAMESPACE")).Get(context.Background(), secretName, v1.GetOptions{}); err != nil {
		return accessToken, err
	} else {
		logger.WithField("func", "getAccessToken").Infof("found access-token with length %d in secret %s", len(secret.Data["access-token"]), secret.Name)
		return string(secret.Data["access-token"]), nil
	}
}

func (a *GitPromotionTriggeredEventHandler) getNextStage(project string, stage string) (nextStage string, err error) {
	stages, err := a.api.StagesV1().GetAllStages(project)
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

func (a *GitPromotionTriggeredEventHandler) getMergedConfiguration(project string, stage, nextstage string, service string) (config model.PromotionConfig) {
	config = readAndMergeResource(config, func() (resource *models.Resource, err error) {
		return a.api.ResourcesV1().GetProjectResource(project, configurationResource)
	})
	config = readAndMergeResource(config, func() (resource *models.Resource, err error) {
		return a.api.ResourcesV1().GetStageResource(project, stage, configurationResource)
	})
	config = readAndMergeResource(config, func() (resource *models.Resource, err error) {
		return a.api.ResourcesV1().GetServiceResource(project, stage, service, configurationResource)
	})

	placeholders := map[string]string{
		"project":   project,
		"stage":     stage,
		"nextstage": nextstage,
		"service":   service,
	}

	config.Spec.Target.Repo = replacePlaceHolders(placeholders, config.Spec.Target.Repo)
	config.Spec.Target.Secret = replacePlaceHolders(placeholders, config.Spec.Target.Secret)
	for i, p := range config.Spec.Paths {
		p.Target = replacePlaceHolders(placeholders, p.Target)
		p.Source = replacePlaceHolders(placeholders, p.Source)
		config.Spec.Paths[i] = p
	}
	return config
}

func replacePlaceHolders(placeholders map[string]string, p *string) (result *string) {
	if p == nil {
		return nil
	}
	current := *p
	for k, v := range placeholders {
		current = strings.Replace(current, fmt.Sprintf("${%s}", k), v, -1)
	}
	return &current
}

func readAndMergeResource(target model.PromotionConfig, getResourceFunc func() (resource *models.Resource, err error)) (ret model.PromotionConfig) {
	ret = target
	resource, err := getResourceFunc()
	if err == api.ResourceNotFoundError {
		return ret
	}
	var newConfig model.PromotionConfig
	if err := yaml.Unmarshal([]byte(resource.ResourceContent), &newConfig); err != nil {
		logger.WithField("func", "readAndMergeResource").
			WithError(err).
			Errorf("could not unmarshall resource file %s => ignoring", *resource.ResourceURI)
	} else {
		if newConfig.Spec.Strategy != nil {
			ret.Spec.Strategy = newConfig.Spec.Strategy
		}
		if newConfig.Spec.Target.Repo != nil {
			ret.Spec.Target.Repo = newConfig.Spec.Target.Repo
		}
		if newConfig.Spec.Target.Secret != nil {
			ret.Spec.Target.Secret = newConfig.Spec.Target.Secret
		}
		if newConfig.Spec.Target.Provider != nil {
			ret.Spec.Target.Provider = newConfig.Spec.Target.Provider
		}
		ret.Spec.Paths = append(target.Spec.Paths, newConfig.Spec.Paths...)
	}
	return ret
}

func toString(str *string) string {
	if str == nil {
		return "<nil>"
	}
	return *str
}
