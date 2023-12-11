package kube

import (
	"fmt"
	u "github.com/dbschenker/vaultpal/internal/testutil"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	cgt "k8s.io/client-go/util/testing"
	"net/http"
	"os"
	"sort"
	"testing"
)

const (
	PATH_LOOKUP_SELF     = "/v1/auth/token/lookup-self"
	PATH_BRO_CONFIG_BASE = "/v1/kv/data/vaultbro/k8s/clusters/"
	PATH_ISSUE_CERT_F    = "/v1/%s/issue/%s"

	CA = `-----BEGIN CERTIFICATE-----
	MIIDMjCCAhqgAwIBAgIUVLaypw/XmXTBBwj1foltDmH9N5YwDQYJKoZIhvcNAQEL
BQAwFTETMBEGA1UEAxMKazhzLnRzYy5zaDAeFw0xOTEwMjMxMDU4NDFaFw0xOTEw
MjQwMjU5MTFaMBUxEzARBgNVBAMTCms4cy50c2Muc2gwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDdN4sk82TzxG94eJ0+JEy/eEk0oo0RepVOn5EfGj9j
KEFxgKBexsqd/nE5ubYq8mbJjKTRRCy6IxRHD8J4rPTGMQ53STkMRhlmWwxhA/4H
HrBsTjh+J7kfCTx8QsYtkEBEhz/QxYoLBapzf+pCcMO1chNEZWhQ+3VDFYd2DQuw
AkjKjK7126S+tn0T45/tNVtvaGCj+ZQQLz/5f7WK9lHj4trVIdQEc2uGR7i7VJvl
kHO0l3KmnrGDfchn0GH6fp8RjWqyvj+a1VZwUxY86XFWhKTCyEpFfXHXLtZgrFSR
0dg3dgaSiCJuZkW7x3wk5lkWNahIhaW27BMi2FPvUgDjAgMBAAGjejB4MA4GA1Ud
DwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSsBSuaUv+niivI
oUtyLkKRf5S99TAfBgNVHSMEGDAWgBSsBSuaUv+niivIoUtyLkKRf5S99TAVBgNV
HREEDjAMggprOHMudHNjLnNoMA0GCSqGSIb3DQEBCwUAA4IBAQALG0V08Fx5Qklv
4xu42lD5fSuNG8xoKkybeOvSxKBRd3drFjJJq6I/vtrv1Te3ySSatSKnQGBAFwEX
wVbsQ2J9/H7q39i1wpZ+UNgSF+4hQzMDzqj44huMm1dfFYQfxYRLNemcPYgbusSQ
HpkNOFlv3E6ZNRhEmwfSzDF1v0kbP0Qo9LtH2M8laorXk58aLJql/F8IwTOXhrSg
eJi+Tlac8LcC2+HBAbC/CpyTTnPMHe3aMXzaszb8SvdYiIB0tUs1fvd/If1ERIhB
SjGc/ouIEbTp2DTOxpiPeYrsaZSDxVTWT4ntYyBl7khzSKFmBY70tWA2jVslO8b1
igmVv+AW
-----END CERTIFICATE-----`
	CERT = `-----BEGIN CERTIFICATE-----
MIIDNjCCAh6gAwIBAgIUFGWx/UoRW5V/mxMnhmZU2Dpyvr4wDQYJKoZIhvcNAQEL
BQAwFTETMBEGA1UEAxMKazhzLnRzYy5zaDAeFw0xOTEwMjMxMDU4NDFaFw0xOTEw
MjMxMTU5MTFaMA8xDTALBgNVBAMTBHJvb3QwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQD18BQZK+R16krahiQihbRY87A2k/yZ9MegATUoTyutCSCm1Uvi
P1GAFTUM+kxE9XmvFH76pfAaf/8Vw7MT1MHhks+4NDPAB5+0IRlPh2gDafJHDClT
OMUmj0NOMn55OLAA6EHNRExmZreQq55d3HUf49jY9sFc/HeDo5hjJRVAX8Wmi5AB
lt5ZYiUuXZbpNj3MXRQYoY7pU/puhA+uNdaFN3fjOwTnOjkBqUsu6zujxDioiMNS
Ql04zrGx33ODUfM3EvZhY6MEN+T2IsO3YbrgTF1JhjhzPFNeEgCIJhT8h8NCqaaC
E+Lb2zELxjJPRjw2n6xYW2AVUY70kaG9Eg/xAgMBAAGjgYMwgYAwDgYDVR0PAQH/
BAQDAgOoMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAdBgNVHQ4EFgQU
gr8kcWmXMsvKjB8Y7EEJhrl9ukcwHwYDVR0jBBgwFoAUrAUrmlL/p4oryKFLci5C
kX+UvfUwDwYDVR0RBAgwBoIEcm9vdDANBgkqhkiG9w0BAQsFAAOCAQEAOrmv47ic
/GG81mroq2ob/PoSn3voCUUxqaz1nkk0sS1ClEcyxKuUx1kq2sy6T5/JLx4PK3Xm
6wRObQYPH1pagENhzOgdUwrxjOHrk+g0vsqPvTp5qzsUegFKcCQWjVf5rvDqHH/5
7x5cK5/roumSa1LqY+au0t1gCXbWO6N4CQ79wP9saB7szGSw54hfJjwFDwZ+L0Xv
qQI43f/dkMX+cFy7jnNLIb6TdYl1X5Js7mPx0RQJryUUkuJPimZmSeU1cg5Zmv1m
EO9PTDFzIydZ+XqoZAd75sS5E8r3ss3WlBNlw9VvVYqmAN4MKRDPCgB4mYgj9WU1
XAq0OI6rercNGQ==
-----END CERTIFICATE-----`
	PRIVATE_KEY = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA9fAUGSvkdepK2oYkIoW0WPOwNpP8mfTHoAE1KE8rrQkgptVL
4j9RgBU1DPpMRPV5rxR++qXwGn//FcOzE9TB4ZLPuDQzwAeftCEZT4doA2nyRwwp
UzjFJo9DTjJ+eTiwAOhBzURMZma3kKueXdx1H+PY2PbBXPx3g6OYYyUVQF/FpouQ
AZbeWWIlLl2W6TY9zF0UGKGO6VP6boQPrjXWhTd34zsE5zo5AalLLus7o8Q4qIjD
UkJdOM6xsd9zg1HzNxL2YWOjBDfk9iLDt2G64ExdSYY4czxTXhIAiCYU/IfDQqmm
ghPi29sxC8YyT0Y8Np+sWFtgFVGO9JGhvRIP8QIDAQABAoIBADe98HA8GI35Snn5
CVuhvlyi7v+PzyL97fkADRJTz2xqszHdClP/UfOb2uhUGtFOagQauyUIU0FOXXyL
XJ1UDZWY9uejPU966uGi1t/FqveLHdSolv070sOImRKyMyQ6ivnJqpBhuIdFJLnv
i/duLkXKGK4kT3NJ7bSycamXEBgEq0GVZyATX3bYfNonx7cDO+1WKtXSwX8MOQhf
mTqRLWU88LH9z/hWUtoscZsfsj5Nd82KoXAY0c/MdPhcg+haT0e8stYuv1pshcQP
mQf4FMWybSA/7fcbiseHb1bUx9cHj6TdxfBM/NV1FL05pQl6nEcXT/JdP7RGIXcw
NrpDLYECgYEA/bsJatVzs3bRphWI8ZKmcUfvkqKQAOl3S06jgAj0ryptiJsvkMPg
5U1ONC0QAv+q97uQHSLxI7QN1BhPtSi8oYd7vM+x6qm7H1XMmk5NnyBoT0C9MPlE
ySPhehSEsV8FkCcqoDFPxFSIfDTzk+Ixou0ZISKEhnCJ1vAY1xIA+jkCgYEA+CMy
2swzuHHaKVynh0SEpjZsVK+1HvqPBkBe08mctB5BeUdP6M//jLUXgBNg5vv/x4JR
N5imS5qi7vVB6Idaw0tbER86+6ahlNfF/onofhfM19ZxTWVUIAt7CCc+1Mea+Js/
anI4J//3J1+xHFJFEO8EeZoxH3OG7AIZWVOWI3kCgYEA1bvGdQ4VhqmidMtTLlug
hXBZaSYzM/F2oiM+K05f/2Y4GojPCp1WRxJVvDHxePUxabm/7itPAgpcU7ue+TW1
oEPmgehbMReFHyJBVgJ79H1yIMCiHiz8OotVFmdOV7N5ljLH/2VKklG7HxXj0UEL
Gvmq33SaOj12f26FHjZ2SFECgYBS/XrBwPA/bRy5HrsNO7Zd3O/odwfNv6FcRuUw
Ukrt1vyw8k/gnshqqBqfBFwxhPDsKkK9pHlh6es6np6XhcWucaKYnGheyEFchbo7
wqYWniEtwxQL/argONbCSFX0VnoXUd0o3eC4SBzCd3fF8CIXYsmNXiu1yC7E+oK9
5H3fiQKBgQDXXvwE7ZEr1BoDyLTuEAp/wmA0P32ADuSYTIVv+WkO/XXyzk2eIoNY
PRKzmecd8I9CyJ4B/533t+yiXGuyjiiQV4KTCUTqUd94lVbQE5WT+5s7p5DCMZnT
mcg3m9SebKWFrZc/yGdwayXpMtbsAkEKm/rgAQeIWPGbsKif9B+hKA==
-----END RSA PRIVATE KEY-----`
)

type mockData struct {
	identity    string
	clusterName string
	aliasName   string
	pkiName     string
	serverURL   string
	issuingCa   string
	cert        string
	privateKey  string
}

func issueCertPath(pki string, role string) string {
	return fmt.Sprintf(PATH_ISSUE_CERT_F, pki, role)
}

func (m *mockData) mockTokenLookupSelf(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Data: map[string]interface{}{
		"display_name": m.identity,
	}}
	u.WriteJsonResponse(t, sec, w)
}

func (m *mockData) mockReadPalConfig(t *testing.T, w http.ResponseWriter, r *http.Request) {
	var sec api.Secret
	w.Header().Set("Content-Type", "application/json")

	if m.aliasName != "" {
		cdata := map[string]interface{}{
			"name":   m.clusterName,
			"alias":  m.aliasName,
			"server": m.serverURL,
		}
		if m.pkiName != "" {
			cdata["pki"] = m.pkiName
		}
		sec = api.Secret{Data: map[string]interface{}{
			"data": cdata,
		}}
	} else {
		sec = api.Secret{Data: map[string]interface{}{
			"data": map[string]interface{}{
				"name":   m.clusterName,
				"pki":    m.pkiName,
				"server": m.serverURL,
			},
		}}
	}
	u.WriteJsonResponse(t, sec, w)
}

func (m *mockData) mockIssueCert(t *testing.T, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sec := api.Secret{Data: map[string]interface{}{
		"issuing_ca":  m.issuingCa,
		"certificate": m.cert,
		"private_key": m.privateKey,
	}}
	u.WriteJsonResponse(t, sec, w)
}

func TestWriteKubeconfigVerify(t *testing.T) {

	tests := []struct {
		Name       string
		serveMocks map[string]u.ServeMockFunc
		WantErr    string
	}{
		{
			Name: "Missing Identity",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF: (&mockData{identity: ""}).mockTokenLookupSelf,
			},
			WantErr: "identity must be not empty/nil",
		},
		{
			Name: "Missing Cluster Name",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "", pkiName: "k8s-pki", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig,
			},
			WantErr: "cluster name must not be empty",
		},
		{
			Name: "Missing Cluster Name - Nil",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{pkiName: "k8s-pki", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig,
			},
			WantErr: "cluster name must not be empty",
		},
		{
			Name: "Missing PKI Name",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", pkiName: "", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig,
			},
			WantErr: "pki must not be empty",
		},
		{
			Name: "Missing PKI Name - Nil",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig,
			},
			WantErr: "pki must not be empty",
		},
		{
			Name: "Missing Server URL",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", pkiName: "k8s-pki", serverURL: ""}).mockReadPalConfig,
			},
			WantErr: "server must not be empty",
		},
		{
			Name: "Missing Server URL - Nil",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", pkiName: "k8s-pki"}).mockReadPalConfig,
			},
			WantErr: "server must not be empty",
		},
		{
			Name: "Alias Missing Server URL",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", aliasName: "knopf", serverURL: ""}).mockReadPalConfig,
			},
			WantErr: "server must not be empty",
		},
		{
			Name: "Alias Missing Server URL - nil",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", aliasName: "knopf"}).mockReadPalConfig,
			},
			WantErr: "server must not be empty",
		},
		{
			Name: "Alias No PKI expected",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:             (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim": (&mockData{clusterName: "jim", aliasName: "knopf", serverURL: "jim-knopf.tsc.sh", pkiName: "jim-pki"}).mockReadPalConfig,
			},
			WantErr: "pki must be empty",
		},
		{
			Name: "Alias Target Missing PKI",
			serveMocks: map[string]u.ServeMockFunc{
				PATH_LOOKUP_SELF:              (&mockData{identity: "pingpong"}).mockTokenLookupSelf,
				PATH_BRO_CONFIG_BASE + "jim":  (&mockData{clusterName: "jim", aliasName: "emma", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig,
				PATH_BRO_CONFIG_BASE + "emma": (&mockData{clusterName: "emma", serverURL: "emma.tsc.sh"}).mockReadPalConfig,
			},
			WantErr: "pki must not be empty",
		},
	}

	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")

	for _, test := range tests {
		vm.ServeMocks = test.serveMocks
		kubeC, err := handleWriteKubeconfig([]byte{}, "jim", "master")
		assert.EqualError(t, err, test.WantErr)
		assert.Nil(t, kubeC)
	}
}

func TestWriteKubeconfigHTTPMock(t *testing.T) {
	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")
	vm.ServeMocks[PATH_LOOKUP_SELF] = (&mockData{identity: "smurf"}).mockTokenLookupSelf
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"jim"] = (&mockData{clusterName: "jim", pkiName: "k8s-pki", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig
	vm.ServeMocks[issueCertPath("k8s-pki", "master")] = (&mockData{issuingCa: CA, cert: CERT, privateKey: PRIVATE_KEY}).mockIssueCert

	kubeC, err := handleWriteKubeconfig([]byte{}, "jim", "master")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(kubeC))
}

func TestWriteKubeconfigAliasHTTPMock(t *testing.T) {
	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")
	vm.ServeMocks[PATH_LOOKUP_SELF] = (&mockData{identity: "smurf"}).mockTokenLookupSelf
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"jim"] = (&mockData{clusterName: "jim", pkiName: "k8s-pki", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"lukas"] = (&mockData{clusterName: "lukas", serverURL: "lukas.tsc.sh", aliasName: "jim"}).mockReadPalConfig
	vm.ServeMocks[issueCertPath("k8s-pki", "master")] = (&mockData{issuingCa: CA, cert: CERT, privateKey: PRIVATE_KEY}).mockIssueCert

	kubeC, err := handleWriteKubeconfig([]byte{}, "lukas", "master")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(kubeC))
}

func TestWriteKubeconfigMultipleHTTPMock(t *testing.T) {
	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")
	vm.ServeMocks[PATH_LOOKUP_SELF] = (&mockData{identity: "smurf"}).mockTokenLookupSelf
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"emma"] = (&mockData{clusterName: "emma", pkiName: "k8s-pki-emma", serverURL: "emma.tsc.sh"}).mockReadPalConfig
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"jim"] = (&mockData{clusterName: "jim", pkiName: "k8s-pki", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"lukas"] = (&mockData{clusterName: "lukas", serverURL: "lukas.tsc.sh", aliasName: "jim"}).mockReadPalConfig
	vm.ServeMocks[issueCertPath("k8s-pki", "master")] = (&mockData{issuingCa: CA, cert: CERT, privateKey: PRIVATE_KEY}).mockIssueCert
	vm.ServeMocks[issueCertPath("k8s-pki-emma", "lokomotive")] = (&mockData{issuingCa: CA, cert: CERT, privateKey: PRIVATE_KEY}).mockIssueCert

	expected := Config{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []ClusterEntry{{
			Name: "lukas",
			Cluster: Cluster{
				Server:                   "lukas.tsc.sh",
				CertificateAuthorityData: StringToBase64String(CA),
			},
		}},
		Contexts: []ContextEntry{
			{
				Name: "lukas",
				Context: Context{
					Cluster:   "lukas",
					Namespace: "master",
					User:      "lukas_smurf",
				},
			},
		},
		Users: []UserEntry{
			{
				Name: "lukas_smurf",
				User: User{
					ClientCertificateData: StringToBase64String(CERT),
					ClientKeyData:         StringToBase64String(PRIVATE_KEY),
				},
			},
		},
		CurrentContext: "lukas",
	}
	kubeC, err := handleWriteKubeconfig([]byte{}, "lukas", "master")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(kubeC))
	assertKubeConfig(t, expected, kubeC)

	kubeC, err = handleWriteKubeconfig([]byte(kubeC), "emma", "lokomotive")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(kubeC))
	expected.Clusters = append(expected.Clusters, ClusterEntry{
		Name: "emma",
		Cluster: Cluster{
			Server:                   "emma.tsc.sh",
			CertificateAuthorityData: StringToBase64String(CA),
		},
	})
	expected.Users = append(expected.Users, UserEntry{
		Name: "emma_smurf",
		User: User{
			ClientCertificateData: StringToBase64String(CERT),
			ClientKeyData:         StringToBase64String(PRIVATE_KEY),
		},
	})
	expected.Contexts = append(expected.Contexts, ContextEntry{
		Name: "emma",
		Context: Context{
			Cluster:   "emma",
			Namespace: "lokomotive",
			User:      "emma_smurf",
		},
	})
	expected.CurrentContext = "emma"
	assertKubeConfig(t, expected, kubeC)

	kubeC, err = handleWriteKubeconfig([]byte(kubeC), "jim", "master")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(kubeC))
	expected.Clusters = append(expected.Clusters, ClusterEntry{
		Name: "jim",
		Cluster: Cluster{
			Server:                   "jim-knopf.tsc.sh",
			CertificateAuthorityData: StringToBase64String(CA),
		},
	})
	expected.Users = append(expected.Users, UserEntry{
		Name: "jim_smurf",
		User: User{
			ClientCertificateData: StringToBase64String(CERT),
			ClientKeyData:         StringToBase64String(PRIVATE_KEY),
		},
	})
	expected.Contexts = append(expected.Contexts, ContextEntry{
		Name: "jim",
		Context: Context{
			Cluster:   "jim",
			Namespace: "master",
			User:      "jim_smurf",
		},
	})
	expected.CurrentContext = "jim"
	assertKubeConfig(t, expected, kubeC)
}

func TestWriteKubeconfigFile(t *testing.T) {
	vm := u.NewVaultServerMock(t)
	defer vm.CloseServer()
	dir, err := cgt.MkTmpdir("kube")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv(api.EnvVaultAddress, vm.Server.URL)
	os.Setenv(api.EnvVaultToken, "1234")
	os.Setenv(ENV_VAULTPAL_KUBECONFIG_FILE, dir+"/bro_test_config")
	broFile := os.Getenv(ENV_VAULTPAL_KUBECONFIG_FILE)

	err = os.Remove(broFile)
	if err != nil {
		t.Logf("vaultpal kube config file did not exist %s", broFile)
	}

	vm.ServeMocks[PATH_LOOKUP_SELF] = (&mockData{identity: "smurf"}).mockTokenLookupSelf
	vm.ServeMocks[PATH_BRO_CONFIG_BASE+"jim"] = (&mockData{clusterName: "jim", pkiName: "k8s-pki", serverURL: "jim-knopf.tsc.sh"}).mockReadPalConfig
	vm.ServeMocks[issueCertPath("k8s-pki", "master")] = (&mockData{issuingCa: CA, cert: CERT, privateKey: PRIVATE_KEY}).mockIssueCert

	err = WriteKubeconfig("jim", "master")
	if err != nil {
		t.Fatal(err)
	}
	cRaw, err := os.ReadFile(broFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := Config{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []ClusterEntry{{
			Name: "jim",
			Cluster: Cluster{
				Server:                   "jim-knopf.tsc.sh",
				CertificateAuthorityData: StringToBase64String(CA),
			},
		}},
		Contexts: []ContextEntry{
			{
				Name: "jim",
				Context: Context{
					Cluster:   "jim",
					Namespace: "master",
					User:      "jim_smurf",
				},
			},
		},
		Users: []UserEntry{
			{
				Name: "jim_smurf",
				User: User{
					ClientCertificateData: StringToBase64String(CERT),
					ClientKeyData:         StringToBase64String(PRIVATE_KEY),
				},
			},
		},
		CurrentContext: "jim",
	}
	assertKubeConfig(t, expected, cRaw)

	// Run again - this will read file and should return same
	err = WriteKubeconfig("jim", "master")
	if err != nil {
		t.Fatal(err)
	}
	cRaw, err = os.ReadFile(broFile)
	if err != nil {
		t.Fatal(err)
	}
	assertKubeConfig(t, expected, cRaw)
}

func TestDeriveNamespaceFromRole(t *testing.T) {
	testNamespace(t, "hase-user", "hase")
	testNamespace(t, "hase-admin", "hase")
	testNamespace(t, "hase", "hase")
}

func testNamespace(t *testing.T, role string, expectedNamespace string) {
	actualNamespace := deriveNamespaceFromRole(role)
	if actualNamespace != expectedNamespace {
		t.Errorf("handler returned unexpected namespace: got %v expected %v", actualNamespace, expectedNamespace)
	}
}

func assertKubeConfig(t *testing.T, want Config, gotRaw []byte) {
	got, err := parseKubeConfig(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	sortEntries(t, &want)
	sortEntries(t, got)
	assert.EqualValues(t, want, *got)
}

func sortEntries(t *testing.T, config *Config) {

	sort.Slice(config.Contexts, func(i, j int) bool {
		return config.Contexts[i].Name < config.Contexts[j].Name
	})
	sort.Slice(config.Users, func(i, j int) bool {
		return config.Users[i].Name < config.Users[j].Name
	})
	sort.Slice(config.Clusters, func(i, j int) bool {
		return config.Clusters[i].Name < config.Clusters[j].Name
	})
}

/**
This tests can be used to run with vault mock to test stuff.

func TestWriteKubeconfig(t *testing.T) {
	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"pki": pki.Factory,
			"kv":  kv2.Factory,
		},
	}
	core, _, rootT := vault.TestCoreUnsealedWithConfig(t, coreConfig)
	ln, addr := vhttp.TestServer(t, core)
	defer ln.Close()
	os.Setenv(api.EnvVaultAddress, addr)
	os.Setenv(api.EnvVaultToken, rootT)

	cl, _ := api.NewClient(&api.Config{
		Address: addr,
	})
	cl.SetToken(rootT)
	err := cl.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Config: api.MountConfigInput{
			DefaultLeaseTTL: "16h",
			MaxLeaseTTL:     "32h",
		},
		Options: map[string]string{
			"version": "2",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mounts, _ := cl.Sys().ListMounts()
	log.Infof("Mount Type %s, Options %s", mounts["kv/"].Type, mounts["kv/"].Options)

	bro := map[string]interface{}{
		"data": map[string]interface{}{
			"name":   "jim",
			"pki":    "k8s-pki",
			"server": "jim-knopf.tsc.sh",
		},
	}
	resp, err := cl.Logical().Write("kv/data/vaultbro/k8s/clusters/kube", bro)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("no return")
	}

	err = cl.Sys().Mount("k8s-pki", &api.MountInput{
		Type: "pki",
		Config: api.MountConfigInput{
			DefaultLeaseTTL: "16h",
			MaxLeaseTTL:     "32h",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err = cl.Logical().Write("k8s-pki/root/generate/internal", map[string]interface{}{
		"common_name": "k8s.tsc.sh",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("no ca provided")
	}

	// Create kube pki role
	_, err = cl.Logical().Write("k8s-pki/roles/master", map[string]interface{}{
		"max_ttl":        "2h",
		"require_cn":     false,
		"allow_any_name": true,
	})

	err = WriteKubeconfig("kube", "master")
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteKubeconfigWithAlias(t *testing.T) {
	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"pki": pki.Factory,
			"kv":  kv2.Factory,
		},
	}
	core, _, rootT := vault.TestCoreUnsealedWithConfig(t, coreConfig)
	ln, addr := vhttp.TestServer(t, core)
	defer ln.Close()
	os.Setenv(api.EnvVaultAddress, addr)
	os.Setenv(api.EnvVaultToken, rootT)

	cl, _ := api.NewClient(&api.Config{
		Address: addr,
	})
	cl.SetToken(rootT)
	err := cl.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Config: api.MountConfigInput{
			DefaultLeaseTTL: "16h",
			MaxLeaseTTL:     "32h",
		},
		Options: map[string]string{
			"version": "2",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mounts, _ := cl.Sys().ListMounts()
	log.Infof("Mount Type %s, Options %s", mounts["kv/"].Type, mounts["kv/"].Options)

	bro := map[string]interface{}{
		"data": map[string]interface{}{
			"name":   "jim",
			"pki":    "k8s-pki",
			"server": "jim-knopf.tsc.sh",
		},
	}
	resp, err := cl.Logical().Write("kv/data/vaultbro/k8s/clusters/jim", bro)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("no return")
	}

	alias := map[string]interface{}{
		"data": map[string]interface{}{
			"name":   "lukas",
			"server": "lukas.tsc.sh",
			"alias":  "jim",
		},
	}
	resp, err = cl.Logical().Write("kv/data/vaultbro/k8s/clusters/lukas", alias)
	if err != nil {
		t.Fatal(err)
	}

	err = cl.Sys().Mount("k8s-pki", &api.MountInput{
		Type: "pki",
		Config: api.MountConfigInput{
			DefaultLeaseTTL: "16h",
			MaxLeaseTTL:     "32h",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err = cl.Logical().Write("k8s-pki/root/generate/internal", map[string]interface{}{
		"common_name": "k8s.tsc.sh",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("no ca provided")
	}

	// Create kube pki role
	_, err = cl.Logical().Write("k8s-pki/roles/master", map[string]interface{}{
		"max_ttl":        "2h",
		"require_cn":     false,
		"allow_any_name": true,
	})

	err = WriteKubeconfig("lukas", "master")
	if err != nil {
		t.Fatal(err)
	}
}
*/
