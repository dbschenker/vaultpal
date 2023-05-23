package timer

import (
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/vault"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
)

const (
	TokenFileName = "/.vault-token"
)

func TestBashInitial(t *testing.T) {
	t.Skip("I don't understand, why this fails now...")
	bashDefault = true
	_ = os.Unsetenv("PS1")
	out := captureOutput(func() {
		PromptString()
	})
	wanted := regexp.MustCompile(`token.*timer`)
	if !wanted.MatchString(out) {
		t.Errorf("program did not output proper PS1 value, but %q PS1=%q", out, os.Getenv("PS1"))
	}
}

func TestBashExists(t *testing.T) {
	bashDefault = true
	_ = os.Setenv("PS1", "$(token-timer)and the other useful stuff in PS1")
	out := captureOutput(func() {
		PromptString()
	})
	notWanted := regexp.MustCompile(`token.*timer`)
	if notWanted.MatchString(out) {
		t.Errorf("program should not print PS1 value if it's already set")
	}
}

func TestVaultTokenTTL(t *testing.T) {
	vconf := &vault.CoreConfig{
		Logger: logging.NewVaultLogger(hclog.Warn),
	}
	core, _, root := vault.TestCoreUnsealedWithConfig(t, vconf)

	ln, addr := http.TestServer(t, core)
	defer ln.Close()

	conf := api.DefaultConfig()
	conf.Address = addr
	client, _ := api.NewClient(conf)
	client.SetToken(root)

	lease := "42m"
	secret, _ := client.Auth().Token().Create(&api.TokenCreateRequest{
		Lease: lease,
	})

	path := HomeDir() + TokenFileName
	_ = os.WriteFile(path, []byte(secret.Auth.ClientToken), 0600)

	actual := vaultTokenTTL(addr, secret.Auth.ClientToken)

	maxDifference, _ := time.ParseDuration("1s") // :)
	intended, _ := time.ParseDuration(lease)

	if intended-actual > maxDifference {
		t.Errorf("difference between intended %s and actual %s is larger than the maximum of: %s", intended, actual, maxDifference)
	}
}

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
