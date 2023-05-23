package timer

import (
	"context"
	"errors"
	"fmt"
	"github.com/dbschenker/vaultpal/timer/cache"
	"math"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/command/config"
)

const (
	UsualMaxTTL    = time.Hour
	Green          = "green"
	Yellow         = "yellow"
	Red            = "red"
	NetworkTimeout = 25 * time.Millisecond
)

var (
	bashDefault = false
)

func Timer(bash bool, query bool, clear bool) {

	if bash {
		PromptString()
		return
	}

	if clear {
		cache.Clear()
		return
	}

	vaultAddr := os.Getenv(api.EnvVaultAddress)
	if vaultAddr == "" {
		os.Exit(0)
	}

	token, err := currentToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get token failed: %e\n", err)
		os.Exit(0)
	}
	if token == "" {
		fmt.Fprintf(os.Stderr, "token empty\n")
		os.Exit(0)
	}

	ttl := vaultTokenTTL(vaultAddr, token)
	if ttl > 0 {
		output(ttl, false, label(vaultAddr), query)
	} else {
		os.Exit(1)
	}
	os.Exit(0)
}

func currentToken() (string, error) {
	tokenHelper, err := config.DefaultTokenHelper()
	if err != nil {
		return "", fmt.Errorf("error getting token helper: %s", err)
	}
	token, err := tokenHelper.Get()
	if err != nil {
		return "", fmt.Errorf("error getting token: %s", err)
	}
	return strings.TrimSpace(token), nil
}

func PromptString() {
	me := fmt.Sprintf("%s %s", os.Args[0], os.Args[1])
	exists := regexp.MustCompile(me)
	if !exists.MatchString(os.Getenv("PS1")) {
		fmt.Printf("PS1=\"\\$(%s)$PS1\"", me)
	}
}

func label(address string) string {
	switch address {
	case "https://vault.x.sh":
		return "[SB]"
	case "https://vault.y.sh":
		return "[NP]"
	case "https://vault.x.run":
		return "[PR]"
	case "https://vault.security.aws.x.net":
		return "[DSB]"
	case "https://vault-p-np.security.aws.x.com":
		return "N "
	case "https://vault-p-pr.security.aws.x.com":
		return "P "
	default:
		return "[??]"
	}
}

func vaultTokenTTL(endpoint string, currentToken string) time.Duration {
	var cached = new(cache.Cache)
	err := cache.ReadCache(cached)
	if err == nil && cached.Address == endpoint && cached.Token == currentToken {
		// use cache and skip expensive vault network access
		now := time.Now()
		passed := now.Sub(cached.Updated)
		newTTL := cached.TTL - passed
		if newTTL > 0 {
			_ = cache.WriteCache(cache.Cache{
				Address: cached.Address,
				Token:   cached.Token,
				Updated: now,
				TTL:     newTTL,
			})
			return newTTL
		}
	}

	if err := verifyNetwork(endpoint); err != nil {
		// no network: no vault
		os.Exit(0)
	}

	// expensive

	client, err := api.NewClient(&api.Config{
		Address: endpoint,
		Timeout: 1300 * time.Millisecond,
	})
	if err != nil {
		os.Exit(1)
	}
	client.SetToken(currentToken)
	t, err := client.Auth().Token().LookupSelf()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "unset your VAULT_ADDR variable, %s can't be reached \n", endpoint)
		}
		return 0
	}
	ttl, err := t.TokenTTL()
	if err != nil {
		return 0
	}

	_ = cache.WriteCache(cache.Cache{
		Address: endpoint,
		Token:   currentToken,
		Updated: time.Now(),
		TTL:     ttl,
	})
	return ttl
}

func verifyNetwork(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	defer cancel()

	_, err = net.DefaultResolver.LookupIPAddr(ctx, u.Hostname())
	if err != nil {
		return err
	}

	return nil
}

func output(ttl time.Duration, carriageReturn bool, label string, query bool) {
	if ttl <= 0 {
		fmt.Printf("")
		return
	}
	factor := math.Floor(ttl.Seconds() / UsualMaxTTL.Seconds() * 100)
	var fmtCR = "\r"
	if !carriageReturn {
		fmtCR = ""
	}
	info := color.New(color.FgGreen)
	warn := color.New(color.FgYellow)
	crit := color.New(color.FgRed)
	msg := fmt.Sprintf("%s%s%02dm ", fmtCR, label, ttl/time.Minute)
	switch {
	case factor >= 50:
		if query {
			fmt.Println(Green)
		} else {
			info.Printf(msg)
		}

	case factor <= 50 && factor >= 10:
		if query {
			fmt.Println(Yellow)
		} else {
			warn.Printf(msg)
		}

	default:
		if query {
			fmt.Println(Red)
		} else {
			crit.Printf(msg)
		}
	}
}