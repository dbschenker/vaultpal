package vault

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/command/config"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"strings"
)

func NewClient() (*api.Client, error) {
	vaultHost := os.Getenv(api.EnvVaultAddress)

	// Env variable VAULT_TOKEN takes precedence, similar to Vault CLI
	// if unset, fallback to ~/.vault-token or external token helper
	token := os.Getenv(api.EnvVaultToken)
	if token == "" {
		tokenHelper, err := config.DefaultTokenHelper()
		if err != nil {
			return nil, fmt.Errorf("error getting token helper: %s", err)
		}
		token, err = tokenHelper.Get()
		if err != nil {
			return nil, fmt.Errorf("error getting token: %s", err)
		}
	}

	_, err := url.ParseRequestURI(vaultHost)
	if err != nil {
		return nil, errors.Wrap(err, "invalid vault address provided. Check environment variable [VAULT_ADDR]")
	}

	client, err := api.NewClient(&api.Config{
		Address: vaultHost,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating vault client")
	}

	client.SetToken(strings.TrimSpace(token))

	return client, nil
}

func GetIdentityName(client *api.Client) (*string, error) {
	self, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return nil, errors.Wrap(err, "error lookup own identity")
	}

	name, ok := self.Data["display_name"]
	if !ok {
		return nil, fmt.Errorf("could no find display name: %v", self.Data)
	}
	ret := name.(string)
	return &ret, nil
}

func GetVerifiedSecretString(secret *api.Secret, dataKey string, denyEmptyString bool) (string, error) {

	if value, exists := secret.Data[dataKey]; exists {
		stValue, ok := value.(string)
		if ok {
			if denyEmptyString && len(stValue) == 0 {
				return "", errors.Errorf("item value of [%s] in secret data must not be empty", dataKey)
			} else {
				return stValue, nil
			}
		} else {
			return "", errors.Errorf("item value of [%s] in secret data cannot be converted to string", dataKey)
		}
	} else {
		return "", errors.Errorf("item [%s] does not exist in secret data", dataKey)
	}
}
