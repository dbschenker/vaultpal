package vault

import (
	"os"
	"testing"
)

func TestVaultTokenEnv(t *testing.T) {
	expect := "s.220vault"
	_ = os.Setenv("VAULT_ADDR", "https://a.b.c") // any valid URL is ok
	_ = os.Setenv("VAULT_TOKEN", expect)
	client, err := NewClient()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if client.Token() != expect {
		t.Errorf("Expected %s got %s", expect, client.Token())
	}
}
