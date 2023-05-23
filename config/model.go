package config

import "time"

type KubeCluster struct {
	Server string `json: "server"`
	Name   string `json: "name"`
	PKI    string `json: "pki"`
	Alias  string `json: "alias",omitempty`
}

// AWSCredentials represents the set of attributes used to authenticate to AWS with a short lived session
type AWSCredentials struct {
	AWSAccessKey     string    `ini:"aws_access_key_id"`
	AWSSecretKey     string    `ini:"aws_secret_access_key"`
	AWSSessionToken  string    `ini:"aws_session_token"`
	AWSSecurityToken string    `ini:"aws_security_token"`
	Expires          time.Time `ini:"x_security_token_expires"`
	Region           string    `ini:"region,omitempty"`
}
