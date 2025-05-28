package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	config2 "github.com/dbschenker/vaultpal/config"
	"github.com/dbschenker/vaultpal/utils"
	"github.com/dbschenker/vaultpal/vault"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

var stsExportStr = `export AWS_ACCESS_KEY_ID=%s
export AWS_SECRET_ACCESS_KEY=%s
export AWS_SESSION_TOKEN=%s`

var stsCmdEnvStr = `set AWS_ACCESS_KEY_ID=%s
set AWS_SECRET_ACCESS_KEY=%s
set AWS_SESSION_TOKEN=%s`

var stsPsEnvStr = `$env:AWS_ACCESS_KEY_ID="%s"
$env:AWS_SECRET_ACCESS_KEY="%s"
$env:AWS_SESSION_TOKEN="%s"`

type creds struct {
	SessionId    string
	SessionKey   string
	SessionToken string
}

const (
	errFederationMarshal   = "failed to marshal federation session"
	errFederationRequest   = "failed to request federation"
	errFederationResponse  = "failed to receive federation response body"
	errFederationUnmarshal = "failed to unmarshal sign-in token"
	defaultEngine          = "aws"
)

func ExportSTSCredentials(engine string, role string) error {

	exportCmd, err := handleExportSTSCreds(engine, role)
	if err != nil {
		// we must make sure, nothing goes to STDOUT
		//log.Error("Failed: ", err)
		return err
	}
	_, _ = os.Stdout.Write([]byte(exportCmd))
	return nil
}

func handleExportSTSCreds(engine string, role string) (string, error) {

	creds, err := getCreds(engine, role)
	if err != nil {
		return "", err
	}
	stsFormatStr := stsExportStr
	if runtime.GOOS == "windows" {
		log.Debugln("OS Detected: Running on Windows!")
		hostShell, err := utils.GetHostShell()
		if err != nil {
			log.Errorln("Failed to get host shell: ", err)
		}
		log.Debugln("Host Shell: ", hostShell)
		if hostShell == "cmd" {
			log.Debugln("Shell Host Detected: Running on CMD!")
			stsFormatStr = stsCmdEnvStr
		} else if hostShell == "powershell" {
			log.Debugln("Shell Host Detected: Running on PowerShell!")
			stsFormatStr = stsPsEnvStr
		} else {
			return "", errors.New("Failed to detect host shell!")
		}
	}
	exportCmd := fmt.Sprintf(stsFormatStr, creds.SessionId, creds.SessionKey, creds.SessionToken)
	return exportCmd, nil
}

func getCreds(engine string, role string) (creds, error) {
	client, err := vault.NewClient()

	if err != nil {
		return creds{}, errors.Wrap(err, "error creating vault api client")
	}

	secret, err := client.Logical().Read(fmt.Sprintf("%s/sts/%s", engine, role))
	if err != nil {
		return creds{}, errors.Wrap(err, "error reading STS credentials from vault")
	}

	if secret == nil {
		return creds{}, fmt.Errorf("error reading STS credentials from vault")
	}

	secretKey, err := vault.GetVerifiedSecretString(secret, "secret_key", true)
	if err != nil {
		return creds{}, err
	}

	accessKey, err := vault.GetVerifiedSecretString(secret, "access_key", true)
	if err != nil {
		return creds{}, err
	}

	securityToken, err := vault.GetVerifiedSecretString(secret, "security_token", true)
	if err != nil {
		return creds{}, err
	}

	session := creds{
		SessionId:    accessKey,
		SessionKey:   secretKey,
		SessionToken: securityToken,
	}

	return session, nil
}

func GenerateConsoleURL(engine string, suppressBrowser bool, role string) error {

	creds, err := getCreds(engine, role)
	if err != nil {
		// we must make sure, nothing goes to STDOUT
		//log.Error("Failed: ", err)
		return err
	}

	signInToken, err := createSignInToken(creds)
	if err != nil {
		return err
	}

	consoleURL := signInURL(signInToken)
	_, _ = os.Stdout.Write([]byte(consoleURL))

	if !suppressBrowser {
		_ = openBrowser(consoleURL)
	}

	return nil
}

func signInURL(signInToken string) string {
	var (
		destination = "https://console.aws.amazon.com/"
		issuer      = os.Getenv(api.EnvVaultAddress)
	)
	if os.Getenv("AWS_REGION") != "" {
		destination = fmt.Sprintf("https://%s.console.aws.amazon.com/", os.Getenv("AWS_REGION"))
	}

	return fmt.Sprintf("https://signin.aws.amazon.com/federation?Action=login&Issuer=%s&Destination=%s&SigninToken=%s\n",
		issuer,
		destination,
		signInToken)

}

func createSignInToken(creds creds) (string, error) {

	session := map[string]string{
		"sessionId":    creds.SessionId,
		"sessionKey":   creds.SessionKey,
		"sessionToken": creds.SessionToken,
	}

	enc, err := json.Marshal(session)

	if err != nil {
		return "", errors.Wrapf(err, errFederationMarshal)
	}

	tokenURL := fmt.Sprintf("https://signin.aws.amazon.com/federation?Action=getSigninToken&Session=%s",
		url.QueryEscape(string(enc)))

	var buf = bytes.NewBuffer(nil)

	res, err := http.Get(tokenURL)

	if err != nil {
		return "", errors.Wrapf(err, errFederationRequest)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if _, err := io.Copy(buf, res.Body); err != nil {
		return "", errors.Wrapf(err, errFederationResponse)
	}

	var body map[string]string

	if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
		return "", errors.Wrapf(err, errFederationUnmarshal)
	}

	return body["SigninToken"], nil
}

func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	} else {
		return nil
	}
}

func WriteAWSCreds(role string, profile string) error {

	filename, err := resolveFilename()
	if err != nil {
		return err
	}

	err = ensureConfigExists(profile)
	if err != nil {
		return errors.Wrap(err, "unable to locate creds file")
	}

	config, err := ini.Load(filename)
	if err != nil {
		return err
	}
	iniProfile, err := config.NewSection(profile)
	if err != nil {
		return err
	}

	creds, err := getCreds(defaultEngine, role)
	if err != nil {
		// we must make sure, nothing goes to STDOUT
		//log.Error("Failed: ", err)
		return err
	}

	test := config2.AWSCredentials{
		AWSAccessKey:     creds.SessionId,
		AWSSecretKey:     creds.SessionKey,
		AWSSessionToken:  creds.SessionToken,
		AWSSecurityToken: creds.SessionToken,
		Expires:          time.Now().Local().Add(1 * time.Hour),
		Region:           "eu-central-1", //maybe drop this
	}

	err = iniProfile.ReflectFrom(&test)
	if err != nil {
		return err
	}

	err = config.SaveTo(filename)
	if err != nil {
		return err
	}

	log.Infof("AWS creds written to: %s \n Use them with `aws --profile %s sts get-caller-identity`", filename, profile)

	return nil
}

func resolveSymlink(filename string) (string, error) {
	sympath, err := filepath.EvalSymlinks(filename)

	// return the un modified filename
	if os.IsNotExist(err) {
		return filename, nil
	}
	if err != nil {
		return "", err
	}

	return sympath, nil
}
func resolveFilename() (string, error) {
	var name string
	var err error
	if runtime.GOOS == "windows" {
		name = path.Join(os.Getenv("USERPROFILE"), ".aws", "credentials")
	} else {
		name, err = homedir.Expand("~/.aws/credentials")
		if err != nil {
			return "", errors.Wrap(err, "user home directory not found")
		}
	}

	// is the filename a symlink?
	name, err = resolveSymlink(name)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve symlink")
	}
	return name, nil
}

func ensureConfigExists(profile string) error {
	filename, err := resolveFilename()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {

			dir := filepath.Dir(filename)

			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}

			// create a base config file
			err = os.WriteFile(filename, []byte("["+profile+"]"), 0600)
			if err != nil {
				return err
			}

		}
		return err
	}
	return nil
}
