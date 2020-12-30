# GitHubBinDl

I need a easy way to get the last binaries from a number of open-source projects
that I follow closely, for example tekton and helm.

I want a easy config file that makes this tool easy to run multiple times,
ether as a cronjob or a manual process.

I trust the community to not perform breaking changes overall and they only provide me with a binary.

Is this a over engineering of a simple script? Most definitely but I think it's fun
and I want a config file to manage what I download in a easy way.

## TODO

- Support auth, this to work around the api limit & private repositories
- Write simple instruction on how to get a GitHub token
- If a current release exist of your cli support to save the old version
- Enable the config file to define how to get the current version of a binary
  - This way we can check to see what current release we have
  - Should be able to get the new release from the github API
  - So we can test before even downloading the cli
- Just write a log instead of download informing if a new version is available
- Be able to look which version of a package you want
- Support .zip instead of only tar.gz
- What build do you want to download? Windows, Darwin, Linux?
- Be able to auto generate new bash/zsh auto-complete
  - We will need to provide a command on how to do this in your config file
  - Also need to define where to store the auto-complete file
- Support github enterprise by being able to define what github endpoint to use
- Support two-factor authentication
- Write some docs
  - Example config
- Auto build go binary and container image
- Not for this project but it would be fun to have a auto-builder for pacman & flatpack of new binary files
- Be able to auto ignore rc/alpha releases
- Verify checksum, the issue here is that github don't store checksum in the github API and there is no standard to store them.
