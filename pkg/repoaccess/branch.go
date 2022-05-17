package repoaccess

import (
	"fmt"
	"github.com/google/go-github/github"
)

func (c *Client) BranchExists(branchName string) (exists bool, err error) {
	if branch, resp, err := c.githubInstance.client.Repositories.GetBranch(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, branchName); err != nil && resp.StatusCode != 404 {
		return false, err
	} else if branch == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func (c *Client) CreateBranch(sourceBranch, targetBranch string) (err error) {
	branch, _, err := c.githubInstance.client.Repositories.GetBranch(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, sourceBranch)
	if err != nil {
		return err
	}
	_, _, err = c.githubInstance.client.Git.CreateRef(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", targetBranch)),
		Object: &github.GitObject{
			SHA: branch.Commit.SHA,
		},
	})
	return err
}

func (c *Client) DeleteBranch(branch string) (err error) {
	if _, err := c.githubInstance.client.Git.DeleteRef(c.githubInstance.context, c.githubInstance.owner, c.githubInstance.repository, fmt.Sprintf("refs/heads/%s", branch)); err != nil {
		return err
	}
	return nil
}
