package commandargs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/topology"
)

func TestUserArgs(t *testing.T) {
	tests := []struct {
		name     string
		shell    Shell
		expected topology.UserArgs
	}{
		{
			name: "all fields populated",
			shell: Shell{
				GitlabUsername:      "jane-doe",
				GitlabKeyID:         "123",
				GitlabKrb5Principal: "jane@EXAMPLE.COM",
			},
			expected: topology.UserArgs{
				Username:      "jane-doe",
				KeyID:         "123",
				Krb5Principal: "jane@EXAMPLE.COM",
			},
		},
		{
			name:     "only username",
			shell:    Shell{GitlabUsername: "jane-doe"},
			expected: topology.UserArgs{Username: "jane-doe"},
		},
		{
			name:     "only key ID",
			shell:    Shell{GitlabKeyID: "123"},
			expected: topology.UserArgs{KeyID: "123"},
		},
		{
			name:     "only krb5 principal",
			shell:    Shell{GitlabKrb5Principal: "jane@EXAMPLE.COM"},
			expected: topology.UserArgs{Krb5Principal: "jane@EXAMPLE.COM"},
		},
		{
			name:     "empty shell",
			shell:    Shell{},
			expected: topology.UserArgs{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.shell.UserArgs())
		})
	}
}
