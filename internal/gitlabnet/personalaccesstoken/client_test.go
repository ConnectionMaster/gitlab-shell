package personalaccesstoken

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitlab-shell/client"
	"gitlab.com/gitlab-org/gitlab-shell/client/testserver"
	"gitlab.com/gitlab-org/gitlab-shell/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet/discover"
)

var (
	requests []testserver.TestRequestHandler
)

func initialize(t *testing.T) {
	requests = []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/personal_access_token",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				var requestBody *RequestBody
				json.Unmarshal(b, &requestBody)

				switch requestBody.KeyId {
				case "0":
					body := map[string]interface{}{
						"success":    true,
						"token":      "aAY1G3YPeemECgUvxuXY",
						"scopes":     [2]string{"read_api", "read_repository"},
						"expires_at": "9001-11-17",
					}
					json.NewEncoder(w).Encode(body)
				case "1":
					body := map[string]interface{}{
						"success": false,
						"message": "missing user",
					}
					json.NewEncoder(w).Encode(body)
				case "2":
					w.WriteHeader(http.StatusForbidden)
					body := &client.ErrorResponse{
						Message: "Not allowed!",
					}
					json.NewEncoder(w).Encode(body)
				case "3":
					w.Write([]byte("{ \"message\": \"broken json!\""))
				case "4":
					w.WriteHeader(http.StatusForbidden)
				}

				if requestBody.UserId == 1 {
					body := map[string]interface{}{
						"success":    true,
						"token":      "YXuxvUgCEmeePY3G1YAa",
						"scopes":     [1]string{"api"},
						"expires_at": nil,
					}
					json.NewEncoder(w).Encode(body)
				}
			},
		},
		{
			Path: "/api/v4/internal/discover",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				body := &discover.Response{
					UserId:   1,
					Username: "jane-doe",
					Name:     "Jane Doe",
				}
				json.NewEncoder(w).Encode(body)
			},
		},
	}
}

func TestGetPersonalAccessTokenByKeyId(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	args := &commandargs.Shell{GitlabKeyId: "0"}
	result, err := client.GetPersonalAccessToken(
		context.Background(), args, "newtoken", &[]string{"read_api", "read_repository"}, "",
	)
	assert.NoError(t, err)
	response := &Response{
		true,
		"aAY1G3YPeemECgUvxuXY",
		[]string{"read_api", "read_repository"},
		"9001-11-17",
		"",
	}
	assert.Equal(t, response, result)
}

func TestGetRecoveryCodesByUsername(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	args := &commandargs.Shell{GitlabUsername: "jane-doe"}
	result, err := client.GetPersonalAccessToken(
		context.Background(), args, "newtoken", &[]string{"api"}, "",
	)
	assert.NoError(t, err)
	response := &Response{true, "YXuxvUgCEmeePY3G1YAa", []string{"api"}, "", ""}
	assert.Equal(t, response, result)
}

func TestMissingUser(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	args := &commandargs.Shell{GitlabKeyId: "1"}
	_, err := client.GetPersonalAccessToken(
		context.Background(), args, "newtoken", &[]string{"api"}, "",
	)
	assert.Equal(t, "missing user", err.Error())
}

func TestErrorResponses(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	testCases := []struct {
		desc          string
		fakeId        string
		expectedError string
	}{
		{
			desc:          "A response with an error message",
			fakeId:        "2",
			expectedError: "Not allowed!",
		},
		{
			desc:          "A response with bad JSON",
			fakeId:        "3",
			expectedError: "Parsing failed",
		},
		{
			desc:          "An error response without message",
			fakeId:        "4",
			expectedError: "Internal API error (403)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			args := &commandargs.Shell{GitlabKeyId: tc.fakeId}
			resp, err := client.GetPersonalAccessToken(
				context.Background(), args, "newtoken", &[]string{"api"}, "",
			)

			assert.EqualError(t, err, tc.expectedError)
			assert.Nil(t, resp)
		})
	}
}

func setup(t *testing.T) (*Client, func()) {
	initialize(t)
	url, cleanup := testserver.StartSocketHttpServer(t, requests)

	client, err := NewClient(&config.Config{GitlabUrl: url})
	require.NoError(t, err)

	return client, cleanup
}