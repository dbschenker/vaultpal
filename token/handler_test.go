package token

import (
	"encoding/json"
	u "github.com/dbschenker/vaultpal/internal/testutil"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/cliconfig"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const (
	PATH_LOOKUP_SELF            = "/v1/auth/token/lookup-self"
	PATH_BASE_CREATE_TOKEN_ROLE = "/v1/auth/token/create"
)

type mockData struct {
	identity    string
	clientToken string
	roleToken   string
}

func (m *mockData) mockTokenLookupSelf(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Data: map[string]interface{}{
		"display_name": m.identity,
	}}
	u.WriteJsonResponse(t, sec, w)
}

func (m *mockData) mockCreateRoleToken(t *testing.T, w http.ResponseWriter, r *http.Request) {
	type roleRequest struct {
		RoleName    string `json:"role_name"`
		TTL         string `json:"ttl"`
		DisplayName string `json:"display_name"`
	}
	var roleR roleRequest
	rbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(rbody, &roleR)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, m.identity, roleR.DisplayName, "expect user identity provided")

	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Auth: &api.SecretAuth{
		ClientToken:      m.roleToken,
		Accessor:         "",
		Policies:         nil,
		TokenPolicies:    nil,
		IdentityPolicies: nil,
		Metadata:         nil,
		Orphan:           false,
		LeaseDuration:    0,
		Renewable:        false,
	}}
	u.WriteJsonResponse(t, sec, w)
}

func TestSwitchTokenRole(t *testing.T) {

	tests := []struct {
		Name            string
		role            string
		serveMocks      map[string]u.ServeMockFunc
		WantErr         string
		WantErrContains string
		WantRoleToken   string
	}{
		{
			Name: "Missing Identity",
			role: "gopher",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF: (&mockData{identity: ""}).mockTokenLookupSelf,
			},
			WantErr: "identity must be not empty/nil",
		},
		{
			Name: "Test normal flow",
			role: "gopher",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:                        (&mockData{identity: "king"}).mockTokenLookupSelf,
				PATH_BASE_CREATE_TOKEN_ROLE + "/gopher": (&mockData{identity: "king", roleToken: "s.1234567890123"}).mockCreateRoleToken,
			},
			WantErr:       "",
			WantRoleToken: "s.1234567890123",
		},
		{
			Name: "Test Client Token invalid",
			role: "gopher",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF: (&u.MockErrorData{HTTPStatus: http.StatusForbidden, Errors: &[]string{"permission denied"}}).MockErrorResponse,
			},
			WantErrContains: "* permission denied",
		},
		{
			Name: "Test Token Role invalid",
			role: "java",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:                      (&mockData{identity: "unknown"}).mockTokenLookupSelf,
				PATH_BASE_CREATE_TOKEN_ROLE + "/java": (&u.MockErrorData{HTTPStatus: http.StatusBadRequest, Errors: &[]string{"unknown role java"}}).MockErrorResponse,
			},
			WantErrContains: "* unknown role java",
		},
	}

	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")

	for _, test := range tests {
		t.Logf("Executing TestCase: %s", test.Name)
		vm.ServeMocks = test.serveMocks
		roleToken, err := createRoleToken(test.role)
		if test.WantErr != "" {
			assert.EqualError(t, err, test.WantErr)
			assert.Nil(t, roleToken)
		} else if test.WantErrContains != "" {
			assert.Contains(t, err.Error(), test.WantErrContains)
			assert.Nil(t, roleToken)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, test.WantRoleToken, roleToken.ClientToken)
		}
	}
}

func TestSwitchTokenRoleWrite(t *testing.T) {

	tests := []struct {
		Name          string
		role          string
		serveMocks    map[string]u.ServeMockFunc
		WantRoleToken string
	}{
		{
			Name: "Test normal flow",
			role: "gopher",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:                        (&mockData{identity: "king"}).mockTokenLookupSelf,
				PATH_BASE_CREATE_TOKEN_ROLE + "/gopher": (&mockData{identity: "king", roleToken: "s.1234567890123"}).mockCreateRoleToken,
			},
			WantRoleToken: "s.1234567890123",
		},
	}

	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")

	for _, test := range tests {
		t.Logf("Executing TestCase: %s", test.Name)
		vm.ServeMocks = test.serveMocks
		err := SwitchRole(test.role)
		assert.Nil(t, err)

		tokenHelper, err := cliconfig.DefaultTokenHelper()
		token, err := tokenHelper.Get()
		assert.Equal(t, test.WantRoleToken, token)
	}
}
