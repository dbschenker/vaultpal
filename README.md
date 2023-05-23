# vaultpal :vulcan_salute:
> { your pal for using vault }

At DB Schenker, we use [HashiCorp Vault](https://www.vaultproject.io/) extensively to automate access to secrets and systems like *Kubernetes Clusters* or the *AWS Cloud*. While the Vault CLI and the [HTTP API](https://developer.hashicorp.com/vault/api-docs) already provide powerful means to access all Vault features, their usage can be cumbersome for end-users who merely want to manage the accesses they need for their daily work.

Vaultpal aims to wrap those function with an easy to use CLI that doesn't deep vault knowledge or excessive configuration. 


## Releases

Check the [releases](./releases) section for binaries suitable for your operating system

Note that Vaultpal has been thoroughly tested on MacOS (`*darwin` binaries) and Linux, but not on Windows.  The binary should execute without issues, but there may be subtle differences in the handling of file locations.
Alternatively, consider [Windows Subsystem for Linux](https://docs.microsoft.com/de-de/windows/wsl/install-win10).


## Setup

1. Extract the go binary, place it on your `$PATH` and make it executable.
    ```bash
      echo $PATH
      chmod a+x /path/to/vaultpal_binary
      cp /path/to/vaultpal_binary /element/of/your/$PATH
    ``` 
   Please note that $PATH output will look like this:
    ```bash
      /Users/your_user/.local/bin:/Users/your_user/.pyenv/plugins/pyenv-virtualenv/shims:
    ```
    That is each element is separated from others by <b>:</b> . So in this example: 
    ```/element/of/your/$PATH``` could be: ```/Users/your_user/.local/bin ``` . You may need to use this command with sudo.
2. Check the installation with:
   ```bash
    vaultpal version
    ```
3. (optional) install completion (bash/zsh). 
    ```bash
    echo 'source <(vaultpal completion bash)' >>~/.bashrc
    ```
4. (optional) define an alias
    ```bash
    echo 'alias vp=vaultpal' >>~/.bashrc
    echo 'complete -F __start_vaultpal vp' >>~/.bashrc
    ```
   
## Usage
- 
- help
   ```
    vaultpal
          { vault~Pal 👍 }
         °- (🕶 ) v1.5.2 -°

   vault~Pal 👍 will help you using vault in your daily work.
   
   Usage:
     vaultpal [flags]
     vaultpal [command]
   
   Available Commands:
     completion  Generates shell completion scripts
     export      Export various types of resources to shell
     help        Help about any command
     switch      Switch between roles
     timer       Display the remaining TTL of your vault token
     version     Print version of vaultpal
     write       Write various types of resources
   
   Flags:
         --config string      config file (default is $HOME/.vaultpal.yaml)
     -h, --help               help for vaultpal
     -v, --verbosity string   Log level (debug, info, warn, error, fatal, panic (default "info")
   
   Use "vaultpal [command] --help" for more information about a command.
   ```

- perform login with vault cli, e.g.:
    ```bash
    vault login -method=oidc
    ```

### Kubeconfig

1. Call vaultpal to create a kubeconfig file
    ```bash
    vaultpal write kubeconfig sandbox master
    ```
    
2. vaultpal will write kubeconfig file in home directory:
    ```bash
    ~/.vaultpal/kube/config
    ``` 
3. Enable usage of kubeconfig file with:
   ```bash
   export KUBECONFIG=~/.vaultpal/kube/config
   ```

### Switch Role

1. Call vaultpal to switch to a token role
    ```bash
    vaultpal switch role k8s-admin
    ```
2. vaultpal will create new token for given role and write it to vault token file
3. Use vault cli with the role token

### Export AWS STS Creds

1. Call vaultpal to create AWS STS credentials with vault
    ```bash
    vaultpal export awssts mytopic-prod-admin
    ```
2. vaultpal will use vault aws secret engine to create AWS STS credentials. The default secret engine path is "aws"
3. The credentials will be printed as bash export commands.

#### Use Alias function

vaultpal provides a bash alias function to wrap the vaultpal command with direct export of the credentials to the current shell.

1. Get the alias function
    ```bash
    vaultpal export awssts -a
    ```
2. Or use it directly in alias definition, e.g.:
    ```bash
    alias vpalsts="$(vaultpal export awssts -a)"
    ```
   Then you will get the credentials exported to current shell and can use it directly, e.g.:
    ```bash
    vpalsts mytopic-prod-admin
    ```

#### Hints

- vaultpal will store a kubeconfig for each cluster with the cluster name as context name
- This enables the usage of different clusters at the same time

## Configuration

The following section describes central configurations, that are required
for vaultpal kubeconfig functions

### Kubeconfig

In order to render kubeconfig files, vaultpal requires meta information about the 
kubernetes cluster. Therefor a cluster configuration object must be stored in vault providing
the required information. The configuration object must be stored in a kv secret engine version 2 at mount path "kv".
The configuration must be accessible for all vaultpal users

Example:
Configuration for a kubernetes cluster called "bibi" must be stored at vault path
`kv/vaultpal/k8s/clusters/bibi`
with data:
```json
{
  "name":   "bibi",
  "pki":    "k8s-bibi-pki-kube",
  "server": "https://api.bibi.mytopic.sh"
}
```
#### Alias

vaultpal supports the definition of an alias to a kubernetes cluster. This is useful if you want to use a generic
endpoint like "int" or "prod" pointing to a cluster.

Example:
Configuration for an alias named "int" pointing to a kubernetes cluster called "bibi" must be stored at vault path
`kv/vaultpal/k8s/clusters/int`
with data:
```json
{
  "name":   "int",
  "alias":  "bibi",
  "server": "https://api.int.mytopic.sh"
}
```
Based on the alias value "bibi", vaultpal will read the configuration for cluster "bibi" in order to render the required certs and keys (pki).  

# Outlook

- further improvements will follow!
- start simple :baby_chick:, use fast :joystick:, enhance :gem:
