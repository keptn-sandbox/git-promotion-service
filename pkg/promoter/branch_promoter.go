package promoter

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"keptn/git-promotion-service/pkg/repoaccess"
	"strings"
)

type BranchPromoter struct {
	client                 repoaccess.Client
	pullRequestTitlePrefix string
}

func NewBranchPromoter(client repoaccess.Client, pullRequestTitlePrefix string) BranchPromoter {
	return BranchPromoter{client: client, pullRequestTitlePrefix: pullRequestTitlePrefix}
}

func (promoter BranchPromoter) Promote(repositoryUrl, fromBranch, toBranch, title, body string) (message string, prLink *string, err error) {
	if newCommits, err := promoter.client.CheckForNewCommits(toBranch, fromBranch); err != nil {
		return "", nil, err
	} else if !newCommits {
		logger.WithField("func", "manageBranchStrategy").Infof("no difference found in repo %s from branch %s to %s", repositoryUrl, fromBranch, toBranch)
		return fmt.Sprintf("no difference between branches %s and %s found => nothing todo", fromBranch, toBranch), nil, nil
	} else if pr, err := promoter.client.GetOpenPullRequest(fromBranch, toBranch); err != nil {
		return "", nil, err
	} else if pr != nil {
		logger.WithField("func", "manageBranchStrategy").Infof("pull request in repo %s from branch %s to %s already open with id %d and title %s", repositoryUrl, fromBranch, toBranch, pr.Number, pr.Title)
		if strings.HasPrefix(pr.Title, promoter.pullRequestTitlePrefix) {
			if err := promoter.client.EditPullRequest(pr, title, body); err != nil {
				return "", nil, err
			}
			logger.WithField("func", "manageBranchStrategy").Infof("updated pull request %d in repo %s from branch %s to %s", pr.Number, repositoryUrl, fromBranch, toBranch)
			return "updated pull request", &pr.URL, nil
		} else {
			return "unmanaged pull request already open", &pr.URL, nil
		}
	} else {
		pr, err := promoter.client.CreatePullRequest(fromBranch, toBranch, title, body)
		if err != nil {
			return message, nil, err
		}
		logger.WithField("func", "manageBranchStrategy").Infof("opened pull request %d in repo %s from branch %s to %s", pr.Number, repositoryUrl, fromBranch, toBranch)
		return "opened pull request", &pr.URL, nil
	}
}
