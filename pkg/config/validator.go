package config

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"keptn/git-promotion-service/pkg/model"
	"net/url"
	"regexp"
	"strings"
)

const githubPathRegexp = "^/[a-zA-Z0-9-]+/[a-zA-Z-_.]+$"

type validator struct {
}

func NewValidator() model.PromotionConfigValidator {
	return validator{}
}

func (v validator) Validate(config model.PromotionConfig) (validationErrrors []string) {
	if config.Spec.Strategy == nil || *config.Spec.Strategy == "" {
		validationErrrors = append(validationErrrors, `"spec.strategy" missing`)
	} else if *config.Spec.Strategy != model.StrategyBranch && *config.Spec.Strategy != model.StrategyFlatPR {
		validationErrrors = append(validationErrrors, fmt.Sprintf(`"spec.strategy" %s invalid`, *config.Spec.Strategy))
	}
	if config.Spec.Target.Secret == nil || *config.Spec.Target.Secret == "" {
		validationErrrors = append(validationErrrors, `"target.secret" missing`)
	}
	if config.Spec.Target.Provider == nil || *config.Spec.Target.Provider == "" {
		validationErrrors = append(validationErrrors, `"target.platform" missing`)
	} else if *config.Spec.Target.Provider != "github" {
		validationErrrors = append(validationErrrors, `target.platform not supported`)
	}
	if config.Spec.Target.Repo == nil || *config.Spec.Target.Repo == "" {
		validationErrrors = append(validationErrrors, `"target.repository" missing`)
	} else {
		u, err := url.Parse(*config.Spec.Target.Repo)
		if err != nil {
			validationErrrors = append(validationErrrors, `"target.repository" is not a valid URL`)
		} else {
			if u.Scheme != "https" || u.Host != "github.com" {
				validationErrrors = append(validationErrrors, `"target.repository" must be a "https" url to a repository on github.com`)
			} else if matched, err := regexp.MatchString(githubPathRegexp, u.Path); err != nil || !matched {
				validationErrrors = append(validationErrrors, `"target.repository" must be a "https" url to a repository on github.com`)
			}
		}
	}
	if config.Spec.Strategy != nil && *config.Spec.Strategy == model.StrategyBranch && len(config.Spec.Paths) > 0 {
		validationErrrors = append(validationErrrors, `no "paths" supported for branch strategy`)
	}
	if config.Spec.Strategy != nil && *config.Spec.Strategy == model.StrategyFlatPR && len(config.Spec.Paths) == 0 {
		validationErrrors = append(validationErrrors, `at least one path is necessary for strategy flat-pr`)
	}
	for i, p := range config.Spec.Paths {
		if p.Target == nil || *p.Target == "" {
			validationErrrors = append(validationErrrors, fmt.Sprintf(`"paths[%d].target" is missing`, i))
		} else {
			//check for targets containing each other (e.g. one target /dev/hello and another /dev/hello/Chart.yaml
			// => this would lead to multiple copy/template operations and errors and is anywayys an inconsistent defininion
			for d, p2 := range config.Spec.Paths {
				if p2.Target != nil && i != d && strings.HasPrefix(*p.Target, *p2.Target) {
					validationErrrors = append(validationErrrors, fmt.Sprintf("paths[%d].target is already included in paths[%d].target", i, d))
				}
			}
		}
		if p.Source != nil && *p.Source == *p.Target {
			validationErrrors = append(validationErrrors, fmt.Sprintf(`"paths[%d].source" is same as target`, i))
		}
	}
	logger.WithField("func", "validateInputEvent").Infof("validation finished with %d validation errors", len(validationErrrors))
	return validationErrrors
}
