package repoaccess

import (
	"github.com/google/go-github/github"
)

type PullRequest struct {
	Number int
	Title  string
	URL    string
}

func (c *Client) GetOpenPullRequest(fromBranch, toBranch string) (pr *PullRequest, err error) {
	prs, _, err := c.githubInstance.client.PullRequests.List(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, &github.PullRequestListOptions{
		Head: fromBranch,
		Base: toBranch,
	})
	if err != nil {
		return pr, err
	}
	if len(prs) == 0 {
		return nil, nil
	}
	pr = &PullRequest{
		Number: *prs[0].Number,
		Title:  *prs[0].Title,
		URL:    *prs[0].HTMLURL,
	}
	return pr, nil
}

func (c *Client) EditPullRequest(pr *PullRequest, title, body string) error {
	if _, _, err := c.githubInstance.client.PullRequests.Edit(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, pr.Number, &github.PullRequest{
		Title: &title,
		Body:  &body,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreatePullRequest(fromBranch, toBranch, title, body string) (pr *PullRequest, err error) {
	ghpr, _, err := c.githubInstance.client.PullRequests.Create(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, &github.NewPullRequest{
		Title: &title,
		Head:  &fromBranch,
		Base:  &toBranch,
		Body:  &body,
	})
	if err != nil {
		return nil, err
	}
	pr = &PullRequest{
		Number: *ghpr.Number,
		Title:  *ghpr.Title,
		URL:    *ghpr.HTMLURL,
	}
	return pr, nil
}
