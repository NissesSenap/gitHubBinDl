# GitHubBinDl

I need a easy way to get the last binaries from a number of open-source projects
that I follow closely, for example tekton and helm.

I want a easy config file that makes this tool easy to run multiple times,
ether as a cronjob or a manual process.

Is this a over engineering of a simple script? Most definitely but I think it's fun
and I want a config file to manage what I download in a easy way.

Hopefully this can be helpful for other developers and system admins that want a easy way of getting
the latest binary download so they can package them for there users.

## Assumptions

The binary that gets downloaded only contain one file that is needed. The archive can contain
multiple files and folders but those will be ignored.

GitHubBinDl don't take any breaking changes in to consideration It just download the binary and unpacks it.

## Getting started

To use GitHubBinDl you don't need to have any github account.

But if you want to be sure that you don't hit the github API requests limit that currently is at 50 an hour as a anonymous user.
If you login you will instead get 5000 requests an hour. Bellow you can find instructions on how to create a github API token.

### Manage config

By default GitHubBinDl uses a data.yaml located in the same folder as your GitHubBinDl binary.
You can override this value by using the environment variable CONFIGFILE and point to another file location.
Or use the `-f` flag when running gitHubBinDl.

If you use a GitHub API token you can store it in two ways,
ether it in the config file or as a environment variable: GITHUBAPIKEY.
If you store it as a environment variable it will take presence over any config in data.yaml.

data.yaml supports the following values:

| Config         | Comment    | Example  | Default |
| -------------- | :----------| :------- | -------: |
| githubAPIkey        | Your github API key         | myAPIkey | "" |
| httpTimeout         | The http timeout in seconds | 5 | 5 |
| httpInsecure        | Allow https without verified certificate | true | false |
| saveLocation        | Where your binary files will be saved | /usr/local/bin | $HOME/gitGubBinDL_\<todays date\> |
| maxFileSize         | The max file size that is allowed to be unpacked from a zip/tar.gz archive in bytes, 1024\*1024\*\<Mb\>| 67108864 | 104857600 |
| notOkCompletionArgs | A list of commands that is not allowed to be provided to the completionArgs| []string{"sudo", "rm"} | []string{"sudo", "rm", "ln", "sed", "awk", "|", "&"} |
| bins                | A list of binaries to download | see bellow | ""|

What values you can have under bin:

| Bins               | Comment | Example | Default |
| ------------------ | :------ | :-------| ------: |
| - cli              | The name of the cli, it have to be exact since it used to match when unpacking the archive | tkn | "" |
| owner              | The github owner |tektoncd | ""|
| repo               | The github repo | cli | ""|
| tag                | A specific tagged release, only support specific version downloads that is tagged. If not defined latest will be used | v0.13.0 | ""|
| match              | How to know which archive to download, GitHubBinDl uses a simple regex match feature | Linux_x86_64 | "" |
| baseURL            | GitHub endpoint, must include a trailing /, should only be used by GitHub enterprise customers | https://api.mygithub.enterprise.com/ | https://api.github.com/ |
| download           | Downloaded package, if not it will just be reported | true | true |
| nonGithubURL       | A non github http server containing tar.gz or .zip fle. If used will ignore any github related config | https://get.helm.sh/helm-v3.4.2-linux-amd64.tar.gz | "" |
| backup             | If true, it will create a copy of the old cli with todays date, example: tkn_2021_01_10 | true | false |
| completionLocation | If set, it will use the newly downloaded bin and generate a completion file, must be the complete path including fileExtension. For more info see [completion generation](#completion-generation) | /tmp/tkn-completion.sh | "" |
| completionArgs     | A list of arguments needed to generate the completion output, one argument per line | - completion - bash | "" |

### Example config

> **Windows** users NOTE that you need to add a file extension

```data.yaml
---
# githubAPIkey: myAPIkey
httpTimeout: 3
httpInsecure: false
saveLocation: /usr/local/bin

bins:
  - cli: tkn
    owner: tektoncd
    repo: cli
    match: Linux_x86_64
    backup: true
    completionLocation: /tmp/tkn-completion.sh
    completionArgs:
      - completion
      - bash
  - cli: tkn.exe
    owner: tektoncd
    repo: cli
    match: Windows_x86_64
    baseURL: https://api.mygithub.enterprise.com/
    download: false
  - cli: kubeseal.exe
    owner: bitnami-labs
    repo: sealed-secrets
    tag: v0.13.1
    match: kubeseal.exe
    baseURL: https://api.github.com/
  - cli: helm
    nonGithubURL: https://get.helm.sh/helm-v3.4.2-linux-amd64.tar.gz
  - cli: helm.exe
    nonGithubURL: https://get.helm.sh/helm-v3.4.2-windows-amd64.zip
```

### Config precedence

The precedence for flag value sources is as follows (highest to lowest):

0. Command line flag value from user
1. Environment variable (if specified)
2. Configuration file value
3. Default value

### Create a GitHub token

It's rather straight forward to generate a Github token, currently I use the UI.

Go to [github.com](github.com) -> settings -> Developer settings -> Personal access tokens -> Generate a new token

For more detailed instructions you can look through this [medium article](https://medium.com/@durgaprasadbudhwani/playing-with-github-api-with-go-github-golang-library-83e28b2ff093).

### Completion generation

There is no verification of the command or the output. The application just takes the full path of completionLocation.
Runs it assuming it got access to run and add the arguments that supplied in completionArgs in that order.

Sadly both tkn & helm both performs a exit 0 even if you provide a incorrect input value for completion.
`helm completion foo == exit 0` even though foo is not a supported option.

After the completion file is generated there is no verification that the file contains any completion code.

## TODO

### priority number 1

- What build do you want to download? Windows, Darwin, Linux?
  - I think this works, but I need to verify on a windows computer CI to the rescue
- Update Makefile to auto-update version nr

### priority number 2

- Be able to auto ignore rc/alpha releases
- validate path and url input in data.yaml
- Just create a json report instead of download informing if a new version is available
  - use the download option in data.yaml
- Write tests both unit and simple e2e
- Not for this project but it would be fun to have a auto-builder for pacman & flatpack of new binary files
- Verify checksum, the issue here is that github don't store checksum in the github API and there is no standard to store them. This won't most likely happen.

## Tests

go test -run Get ./...
