// Package gitauditevent handles Git audit events for GitLab.
package gitauditevent

import (
	"context"
	"fmt"

	pb "gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitlab-shell/v14/client"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/gitlabnet"
)

const uri = "/api/v4/internal/shellhorse/git_audit_event"

// Client handles communication with the GitLab audit event API.
type Client struct {
	config *config.Config
	client *client.GitlabNetClient
}

// NewClient creates a new Client for sending audit events.
func NewClient(config *config.Config) (*Client, error) {
	client, err := gitlabnet.GetClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating http client: %w", err)
	}

	return &Client{config: config, client: client}, nil
}

// Request represents the data for a Git audit event.
type Request struct {
	Action        commandargs.CommandType           `json:"action"`
	Protocol      string                            `json:"protocol"`
	Repo          string                            `json:"gl_repository"`
	Username      string                            `json:"username"`
	PackfileStats *pb.PackfileNegotiationStatistics `json:"packfile_stats,omitempty"`
	CheckIP       string                            `json:"check_ip,omitempty"`
	Changes       string                            `json:"changes"`
}

// Audit sends an audit event to the GitLab API.
func (c *Client) Audit(ctx context.Context, username string, args *commandargs.Shell, repo string, packfileStats *pb.PackfileNegotiationStatistics) error {
	request := &Request{
		Action:        args.CommandType,
		Repo:          repo,
		Protocol:      "ssh",
		Username:      username,
		PackfileStats: packfileStats,
		CheckIP:       gitlabnet.ParseIP(args.Env.RemoteAddr),
		Changes:       "_any",
	}

	response, err := c.client.Post(ctx, uri, request)
	if err != nil {
		return err
	}
	defer response.Body.Close() //nolint:errcheck

	return nil
}
