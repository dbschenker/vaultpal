package aws

import (
	"fmt"
	u "github.com/dbschenker/vaultpal/internal/testutil"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

const (
	PATH_READ_STS_CREDS = "/v1/%s/sts/%s"
)

type mockData struct {
	engine         string
	role           string
	access_key     string
	secret_key     string
	security_token string
}

func (m *mockData) mockReadSTSCreds(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Data: map[string]interface{}{
		"access_key":     m.access_key,
		"secret_key":     m.secret_key,
		"security_token": m.security_token,
	}}
	u.WriteJsonResponse(t, sec, w)
}

func (m *mockData) mockReadWrongCreds(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Data: map[string]interface{}{
		"mop": "map",
	}}
	u.WriteJsonResponse(t, sec, w)
}

func (m *mockData) mockReadSTSCredsSkipEmpty(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Data: map[string]interface{}{}}

	if m.access_key != "" {
		sec.Data["access_key"] = m.access_key
	}
	if m.secret_key != "" {
		sec.Data["secret_key"] = m.secret_key
	}
	if m.security_token != "" {
		sec.Data["security_token"] = m.security_token
	}
	u.WriteJsonResponse(t, sec, w)
}

var (
	mockSuccData          = mockData{engine: "aws", role: "gopher-vpc-manager", access_key: "ACCEZZZZZ123133", secret_key: "SECKEYZZZZ2313112", security_token: "SECTOKENZZZ"}
	mockMissSecretKeyData = mockData{engine: "aws", role: "gopher-vpc-manager", access_key: "ACCEZZZZZ123133", secret_key: "", security_token: "SECTOKENZZZ"}
	mockMissSecTokenData  = mockData{engine: "aws", role: "gopher-vpc-manager", access_key: "ACCEZZZZZ123133", secret_key: "SECKEYZZZZ2313112", security_token: ""}
	mockMissAccKeyData    = mockData{engine: "aws", role: "gopher-vpc-manager", access_key: "", secret_key: "SECKEYZZZZ2313112", security_token: "SECTOKENZZZ"}
	tests                 = []struct {
		Name            string
		MockData        mockData
		serveMocks      map[string]u.ServeMockFunc
		WantErr         string
		WantErrContains string
		WantExportCmd   bool
	}{
		{
			Name:     "Test normal flow",
			MockData: mockSuccData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockSuccData.engine, mockSuccData.role): (mockSuccData).mockReadSTSCreds,
			},
			WantExportCmd: true,
		},
		{
			Name:     "Test invalid role flow",
			MockData: mockSuccData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockSuccData.engine, mockSuccData.role): (&u.MockErrorData{HTTPStatus: http.StatusBadRequest, Errors: &[]string{"Role not found"}}).MockErrorResponse,
			},
			WantExportCmd:   false,
			WantErrContains: "Role not found",
		},
		{
			Name:     "Test empty secret key flow",
			MockData: mockMissSecretKeyData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockMissSecretKeyData.engine, mockMissSecretKeyData.role): (mockMissSecretKeyData).mockReadSTSCreds,
			},
			WantExportCmd: false,
			WantErr:       "item value of [secret_key] in secret data must not be empty",
		},
		{
			Name:     "Test wrong secrets flow",
			MockData: mockMissSecretKeyData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockMissSecretKeyData.engine, mockMissSecretKeyData.role): (mockMissSecretKeyData).mockReadWrongCreds,
			},
			WantExportCmd:   false,
			WantErrContains: "does not exist in secret data",
		},
		{
			Name:     "Test missing access key secret flow",
			MockData: mockMissAccKeyData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockMissAccKeyData.engine, mockMissAccKeyData.role): (mockMissAccKeyData).mockReadSTSCredsSkipEmpty,
			},
			WantExportCmd: false,
			WantErr:       "item [access_key] does not exist in secret data",
		},
		{
			Name:     "Test missing secret key secret flow",
			MockData: mockMissSecretKeyData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockMissSecretKeyData.engine, mockMissSecretKeyData.role): (mockMissSecretKeyData).mockReadSTSCredsSkipEmpty,
			},
			WantExportCmd: false,
			WantErr:       "item [secret_key] does not exist in secret data",
		},
		{
			Name:     "Test missing security token secret flow",
			MockData: mockMissAccKeyData,
			serveMocks: map[string]u.ServeMockFunc{
				fmt.Sprintf(PATH_READ_STS_CREDS, mockMissSecTokenData.engine, mockMissSecTokenData.role): (mockMissSecTokenData).mockReadSTSCredsSkipEmpty,
			},
			WantExportCmd: false,
			WantErr:       "item [security_token] does not exist in secret data",
		},
	}
)

func TestExportSTSCreds(t *testing.T) {

	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")

	for _, test := range tests {
		t.Logf("Executing TestCase: %s", test.Name)
		vm.ServeMocks = test.serveMocks
		err := ExportSTSCredentials(test.MockData.engine, test.MockData.role)
		if test.WantErr != "" {
			assert.EqualError(t, err, test.WantErr)
		} else if test.WantErrContains != "" {
			assert.Contains(t, err.Error(), test.WantErrContains)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestReadSTSCreds(t *testing.T) {

	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")

	for _, test := range tests {
		t.Logf("Executing TestCase: %s", test.Name)
		vm.ServeMocks = test.serveMocks
		exportCmd, err := handleExportSTSCreds(test.MockData.engine, test.MockData.role)
		if test.WantErr != "" {
			assert.EqualError(t, err, test.WantErr)
			assert.Empty(t, exportCmd)
		} else if test.WantErrContains != "" {
			assert.Contains(t, err.Error(), test.WantErrContains)
			assert.Empty(t, exportCmd)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, buildWantExportCmd(test.MockData), exportCmd)
		}
	}
}

func buildWantExportCmd(data mockData) string {
	return fmt.Sprintf(`export AWS_ACCESS_KEY_ID=%s
export AWS_SECRET_ACCESS_KEY=%s
export AWS_SESSION_TOKEN=%s`, data.access_key, data.secret_key, data.security_token)
}
