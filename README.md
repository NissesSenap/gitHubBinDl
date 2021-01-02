# GitHubBinDl

I need a easy way to get the last binaries from a number of open-source projects
that I follow closely, for example tekton and helm.

I want a easy config file that makes this tool easy to run multiple times,
ether as a cronjob or a manual process.

I trust the community to not perform breaking changes overall and they only provide me with a binary.

Is this a over engineering of a simple script? Most definitely but I think it's fun
and I want a config file to manage what I download in a easy way.

Hopefully this can be helpful for other developers and system admins that want a easy way of getting
the latest binary download so they can package them for there users.

## TODO

### priority number 1

- Fix match regexp feature
- If a current release exist of your cli support to save the old version
- Be able to look which version of a package you want
- Add cli option for configfile + version output
- Write simple instruction on how to get a GitHub token
- What build do you want to download? Windows, Darwin, Linux?
  - I think this works, but I need to verify on a windows computer CI to the rescue
- Auto build go binary for linux, darwin and windows
- Write some docs
  - Example config

### priority number 2

- Be able to auto generate new bash/zsh auto-complete
  - We will need to provide a command on how to do this in your config file
  - Also need to define where to store the auto-complete file
- Be able to auto ignore rc/alpha releases
- validate path and url input in data.yaml
- Just create a json report instead of download informing if a new version is available
- Support github enterprise by being able to define what github endpoint to use
- Support two-factor authentication
- Not for this project but it would be fun to have a auto-builder for pacman & flatpack of new binary files
- Verify checksum, the issue here is that github don't store checksum in the github API and there is no standard to store them. This won't most likely happen.
