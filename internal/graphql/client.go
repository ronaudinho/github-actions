// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/sethvargo/go-githubactions"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/FerretDB/github-actions/internal"
)

// Client adds convenience methods to GitHub GraphQL client.
type Client struct {
	*githubv4.Client
	action *githubactions.Action
}

// NewClient returns Client instance with an access token provided from GitHub Actions.
// The token to access API must be provided in the environment variable named `tokenVar`.
func NewClient(ctx context.Context, action *githubactions.Action, tokenVar string) *Client {
	token := action.Getenv(tokenVar)
	if token == "" {
		action.Fatalf("%s is not set.", tokenVar)
		return nil
	}

	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
			Base: internal.NewTransport(action),
		},
	}
	c := githubv4.NewClient(httpClient)

	// check that the client is able to make queries,
	// for that we call a simple rate limit query
	var rl struct {
		RateLimit struct {
			Cost      githubv4.Int
			Limit     githubv4.Int
			Remaining githubv4.Int
			ResetAt   githubv4.DateTime
		}
	}

	if err := c.Query(ctx, &rl, nil); err != nil {
		action.Fatalf("Failed to query rate limit: %s.", err)
		return nil
	}

	action.Debugf(
		"Rate limit remaining: %d, reset at: %s.",
		rl.RateLimit.Remaining, rl.RateLimit.ResetAt.Format(time.RFC3339),
	)

	return &Client{
		Client: c,
		action: action,
	}
}
