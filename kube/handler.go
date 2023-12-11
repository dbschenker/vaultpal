package kube

import (
	"github.com/dbschenker/vaultpal/config"
	"github.com/dbschenker/vaultpal/vault"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"strings"
)

const ENV_VAULTPAL_KUBECONFIG_FILE = "VAULTPAL_KUBECONFIG_FILE"

func contextEntryMap(c []ContextEntry) map[string]ContextEntry {
	cm := map[string]ContextEntry{}

	for _, ce := range c {
		cm[ce.Name] = ce
	}

	return cm
}

func contextEntrySlice(m map[string]ContextEntry) []ContextEntry {
	sl := make([]ContextEntry, len(m))
	idx := 0

	for _, v := range m {
		sl[idx] = v
		idx++
	}

	return sl
}

func clusterEntryMap(c []ClusterEntry) map[string]ClusterEntry {
	cm := map[string]ClusterEntry{}

	for _, ce := range c {
		cm[ce.Name] = ce
	}

	return cm
}

func clusterEntrySlice(m map[string]ClusterEntry) []ClusterEntry {
	sl := make([]ClusterEntry, len(m))
	idx := 0

	for _, v := range m {
		sl[idx] = v
		idx++
	}

	return sl
}

func userEntryMap(u []UserEntry) map[string]UserEntry {
	um := map[string]UserEntry{}

	for _, ue := range u {
		um[ue.Name] = ue
	}

	return um
}

func userEntrySlice(m map[string]UserEntry) []UserEntry {
	sl := make([]UserEntry, len(m))
	idx := 0

	for _, v := range m {
		sl[idx] = v
		idx++
	}

	return sl
}

func getPalKubeConfig(client *api.Client, cluster string) config.KubeCluster {
	vaultpath := "kv/data/vaultbro/k8s/clusters/" + cluster
	k8s, err := client.Logical().Read(vaultpath)
	if err != nil {
		log.Fatal("error reading vaultpal config entry for cluster:  " + err.Error())
	}

	if k8s == nil {
		log.Fatalf("Cluster %s is undefined", cluster)
	}

	cf := config.KubeCluster{}
	mapstructure.Decode(k8s.Data["data"], &cf)
	if cf.Name == "" {
		log.Errorf("Vaultpath %s exists but contains no config for cluster %s", vaultpath, cluster)
	}
	return cf
}

func handleWriteKubeconfig(kconfig []byte, cluster string, role string) ([]byte, error) {
	namespace := deriveNamespaceFromRole(role)
	log.WithFields(log.Fields{
		"Cluster":   cluster,
		"Role":      role,
		"Namespace": namespace,
	}).Info("write a kubeconfig for")
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

	cf := getPalKubeConfig(client, cluster)

	err = verifyPalKubeConfig(cf)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Cluster":   cf.Name,
		"PKI":       cf.PKI,
		"APIServer": cf.Server,
		"Alias":     cf.Alias,
	}).Info("using k8s definition")

	pki := cf.PKI
	if cf.Alias != "" {
		cfAlias := getPalKubeConfig(client, cf.Alias)
		err = verifyPalKubeConfig(cfAlias)
		if err != nil {
			return nil, err
		}
		pki = cfAlias.PKI
		log.WithFields(log.Fields{
			"Alias": cf.Alias,
			"PKI":   cfAlias.PKI,
		}).Infof("[%s] is an alias pointing to [%s]", cf.Name, cfAlias.Name)
	}

	userName := cf.Name + "_" + *user

	secret, err := client.Logical().Write(pki+"/issue/"+role, map[string]interface{}{
		"common_name": *user,
		"ttl":         "3600",
	})
	if err != nil {
		log.Fatal("error creating client key for cluster: " + err.Error())
	}

	clusterE := ClusterEntry{
		Name: cf.Name,
		Cluster: Cluster{
			Server:                   cf.Server,
			CertificateAuthorityData: StringToBase64String(secret.Data["issuing_ca"].(string)),
		},
	}
	contextE := ContextEntry{
		Name: cf.Name,
		Context: Context{
			Cluster:   cf.Name,
			Namespace: namespace,
			User:      userName,
		},
	}
	userE := UserEntry{
		Name: userName,
		User: User{
			ClientCertificateData: StringToBase64String(secret.Data["certificate"].(string)),
			ClientKeyData:         StringToBase64String(secret.Data["private_key"].(string)),
		},
	}

	k8 := Config{
		ApiVersion:     "v1",
		Kind:           "Config",
		CurrentContext: cf.Name,
		Clusters:       []ClusterEntry{},
		Contexts:       []ContextEntry{},
		Users:          []UserEntry{},
	}

	oldk8, err := parseKubeConfig(kconfig)
	if err != nil {
		return nil, err
	}

	contextM := contextEntryMap(oldk8.Contexts)
	contextM[contextE.Name] = contextE
	clusterM := clusterEntryMap(oldk8.Clusters)
	clusterM[clusterE.Name] = clusterE
	userM := userEntryMap(oldk8.Users)
	userM[userE.Name] = userE
	k8.Contexts = contextEntrySlice(contextM)
	k8.Clusters = clusterEntrySlice(clusterM)
	k8.Users = userEntrySlice(userM)

	out, err := yaml.Marshal(k8)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal kubeconfig file")
	}

	log.WithFields(log.Fields{
		"Cluster": cf.Name,
		"API":     cf.Server,
		"Role":    role,
	}).Info("Let's kube 🛀")

	return out, nil
}

func verifyPalKubeConfig(cluster config.KubeCluster) error {

	if cluster.Name == "" {
		return errors.New("cluster name must not be empty")
	}
	if cluster.Server == "" {
		return errors.New("server must not be empty")
	}
	if cluster.Alias == "" {
		if cluster.PKI == "" {
			return errors.New("pki must not be empty")
		}
	} else {
		if cluster.PKI != "" {
			return errors.New("pki must be empty")
		}
	}
	return nil
}

func ensurePalKubeConfigFile() (string, error) {
	kubeconfigFile := ""

	if envPalKFile := os.Getenv(ENV_VAULTPAL_KUBECONFIG_FILE); envPalKFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}

		kubeconfigPath := home + "/.vaultpal/kube"
		kubeconfigFile = kubeconfigPath + "/config"

		// Make sure pal kube dir exists
		err = os.MkdirAll(kubeconfigPath, 0740)
		if err != nil {
			log.WithError(err).Fatal("cannot create dir")
		}
	} else {
		kubeconfigFile = envPalKFile
	}

	return kubeconfigFile, nil
}

func WriteKubeconfig(cluster string, role string) error {

	kubeconfigFile, err := ensurePalKubeConfigFile()
	if err != nil {
		return err
	}

	err = createPalKubeConfigFile(kubeconfigFile)
	if err != nil {
		return err
	}

	kubeConfigR, err := kubeConfigReader(kubeconfigFile)
	if err != nil {
		kubeConfigR.Close()
		return err
	}

	kubeConfigRaw, err := readKubeConfigRaw(kubeConfigR)
	if err != nil {
		kubeConfigR.Close()
		return err
	}
	kubeConfigR.Close()

	newKubeConfig, err := handleWriteKubeconfig(kubeConfigRaw, cluster, role)
	if err != nil {
		return err
	}

	err = os.WriteFile(kubeconfigFile, newKubeConfig, 0600)
	if err != nil {
		return errors.Wrapf(err, "cannot write kubeconfig to [%s]", kubeconfigFile)
	}

	log.Infof("Enable kubeconfig with: KUBECONFIG=%s", kubeconfigFile)

	return nil
}

func createPalKubeConfigFile(file string) error {
	if _, err := os.Stat(file); err == nil {
		return nil
	} else if os.IsNotExist(err) {
		log.Info("create empty bro kube config")
		err = os.WriteFile(file, []byte{}, 0600)
		if err != nil {
			return errors.Wrapf(err, "cannot write kubeconfig to [%s]", file)
		}
		return nil
	} else {
		return err
	}
}

func kubeConfigReader(configFile string) (*os.File, error) {
	kconfigReader, err := os.Open(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read existing kube config [%s]", configFile)
	}
	return kconfigReader, nil
}

func readKubeConfigRaw(configFile io.Reader) ([]byte, error) {
	yamlRaw, err := io.ReadAll(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read existing kube config [%s]", configFile)
	}
	return yamlRaw, nil
}

func parseKubeConfig(configRaw []byte) (*Config, error) {
	kconfig := &Config{}

	err := yaml.Unmarshal(configRaw, kconfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal kube config")
	}

	return kconfig, nil
}

// Traditionally role and namespace have been treated as the same value, e.g. "ttb-int"
// However with the advent of different roles types for the same topic/environment such as ttb-int-user
// We need to strip those suffixes to ensure that the default kubeconfig points to a namespace that actually exists
// Similar to -user we may introduce -admin and additional more fine grained suffixed in future
func deriveNamespaceFromRole(role string) (namespace string) {
	roleSuffixes := [...]string{"-user", "-admin"} // also support -admin, even though it's currently not used
	for _, suffix := range roleSuffixes {
		if strings.HasSuffix(role, suffix) {
			return role[0 : len(role)-len(suffix)]
		}
	}
	return role // traditional behaviour (role == namespace): if no known suffix matches, just return the role input
}
