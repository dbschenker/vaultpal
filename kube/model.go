package kube

import "encoding/base64"

type Cluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
}

type ClusterEntry struct {
	Name    string  `yaml:"name"`
	Cluster Cluster `yaml:"cluster"`
}

type ContextEntry struct {
	Name    string  `yaml:"name"`
	Context Context `yaml:"context"`
}

type Context struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	User      string `yaml:"user"`
}

type UserEntry struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

type User struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

type Config struct {
	ApiVersion     string         `yaml:"apiVersion"`
	Kind           string         `yaml:"kind"`
	Clusters       []ClusterEntry `yaml:"clusters"`
	Contexts       []ContextEntry `yaml:"contexts"`
	Users          []UserEntry    `yaml:"users"`
	CurrentContext string         `yaml:"current-context"`
}

func StringToBase64String(s string) string {
	return base64.StdEncoding.EncodeToString(([]byte(s)))
}
