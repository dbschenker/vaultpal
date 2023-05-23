# 1.2.0 (5-oct-2021)
### BUG FIXES
- remove cache to fix weired docker/golang vendoring bug


# 1.2.0 (5-oct-2021)
### FEATURES
- Token Timer
  - `vaultpal timer` to display the TTL of your current token e.g. form your shell prompt (generate with `vaultpal timer --bash`)

### IMPROVEMENTS
- take potential vault token helpers into account for token read/write

# 1.0.0 (8-sep-2020)
### FEATURES
- Export of AWS credentials to the default `~/.aws/credentials` file
  - store them for different `AWS_PROLFIL`s
  - example:
  ```bash
  $ vaultpal write awscreds topic_owner_tsc default
  INFO[0001] AWS creds written to: /home/pschu/.aws.vaultpal/credentials
   Use them with `aws --profile default sts get-caller-identity`
  ```


# 0.7.0 (24.01.2020)
### FEATURES
- Enable generation of an AWS web console sign-in URL with vault
  - opens it the default browser (can be suppressed with "-s" flag)
  - prints the URL to stdout for sharing and scripting

### IMPROVEMENTS
- Minor refactorings 

# 0.6.0 (15.01.2020)
### FEATURES
- Enable generation of AWS STS credentials with vault
  - provides an alias function for bash
  - prints the export commands to use the AWS Credentials (ENV)
- Provide bash & zsh completion
### IMPROVEMENTS
### BUG FIXES

# 0.5.0 (26.11.2019)
### FEATURES
- Switch to a token role
### IMPROVEMENTS
### BUG FIXES

# 0.4.5 (26.11.2019)
### FEATURES
- Added version cmd and version info
### IMPROVEMENTS
### BUG FIXES

# 0.4.1 (24.10.2019)
### FEATURES
### IMPROVEMENTS
- Added flag to set the log level
    - Run with `vaultpal -v warn ...` to get less output
### BUG FIXES

# 0.4.0 (23.10.2019)
### FEATURES
- Added multi cluster support
  - Defines a context in kubeconfig for each cluster that is used
  - All clusters are defined in one kubeconfig
- Enable definition of an alias to a cluster
  - Use a generic endpoint like "int", "fat" etc. pointing to a cluster
  - This is required to support transparent switch of clusters for users
### IMPROVEMENTS
### BUG FIXES
