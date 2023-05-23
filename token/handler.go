package token

import (
	"fmt"
	"github.com/dbschenker/vaultpal/vault"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/command/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func SwitchRole(role string) error {
	log.WithFields(log.Fields{
		"role": role,
	}).Info("switch to token for ")

	roleTokenAuth, err := createRoleToken(role)
	if err != nil {
		return errors.Wrap(err, "cannot create role token")
	}

	tokenHelper, err := config.DefaultTokenHelper()
	if err != nil {
		return fmt.Errorf("error getting token helper: %s", err)
	}
	err = tokenHelper.Store(roleTokenAuth.ClientToken)
	if err != nil {
		return fmt.Errorf("error update token: %s", err)
	}

	return nil
}

func createRoleToken(role string) (*api.SecretAuth, error) {

	client, err := vault.NewClient()
	if err != nil {
		log.Fatal("error creating vault api client: " + err.Error())
	}

	user, err := vault.GetIdentityName(client)
	if err != nil {
		return nil, errors.Wrap(err, "error getting own identity")
	} else if user == nil || *user == "" {
		return nil, errors.New("identity must be not empty/nil")
	}

	log.WithFields(log.Fields{
		"identity": *user,
	}).Info("got your identity")

	secret, err := client.Logical().Write("auth/token/create/"+role, map[string]interface{}{
		"role_name":    role,
		"ttl":          "1h",
		"display_name": user,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating role token")
	}

	log.WithFields(log.Fields{
		"policies": secret.Auth.TokenPolicies,
	}).Info("bro got a role token for you")

	return secret.Auth, nil
}
