# Quick start

[简体中文](https://github.com/jenkins-zh/jenkins-cli/blob/master/README-zh.md)

## Jenkins CLI


[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/jenkins-zh/jenkins-cli)
[![](https://goreportcard.com/badge/jenkins-zh/jenkins-cli)](https://goreportcard.com/report/jenkins-zh/jenkins-cli)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/jenkins-zh/jenkins-cli)
[![codecov](https://codecov.io/gh/jenkins-zh/jenkins-cli/branch/master/graph/badge.svg?token=XS8g2CjdNL)](https://codecov.io/gh/jenkins-zh/jenkins-cli)
[![Contributors](https://img.shields.io/github/contributors/jenkins-zh/jenkins-cli.svg)](https://github.com/jenkins-zh/jenkins-cli/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/jenkins-zh/jenkins-cli.svg?label=release)](https://github.com/jenkins-zh/jenkins-cli/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/jenkins-zh/jenkins-cli/total)
[![Docker Pulls](https://img.shields.io/docker/pulls/jenkinszh/jcli.svg)](https://hub.docker.com/r/jenkinszh/jcli/tags)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/jenkins-zh/jenkins-cli)
[![HitCount](http://hits.dwyl.com/jenkins-zh/jenkins-cli.svg)](http://hits.dwyl.com/jenkins-zh/jenkins-cli)

Jenkins CLI allows you manage your Jenkins in an easy way. No matter if you're a plugin developer, administrator or just a regular user, it is made for you!

- [Features](#features)
- [Get it](#get-it)
- [Plugins](#plugins)
- [Contribution](#contribution)
- [License](#license)


## Features

* Multiple Jenkins support
* Plugins management \(list, search, install, upload\)
* Job management \(search, build, log\)
* Configuration as Code support
* Open your Jenkins with a browser
* Restart your Jenkins
* Connection with proxy support

## Get it

> 이 저장소는 [jenkins-zh/jenkins-cli](https://github.com/jenkins-zh/jenkins-cli) 의 fork 입니다.
> 원본 대비 변경: 플러그인 없이 `job search` (설정 `search: plugin | api`, 기본 `api`), `폴더/잡` 형식 job 이름 지원

[Homebrew](https://brew.sh/) (macOS, Linux) 로만 배포합니다.

```shell
brew install rody-vx/tap/jcli
```

| 업그레이드 | 삭제 |
|---|---|
| `brew upgrade jcli` | `brew uninstall jcli` |

공식 `jenkins-cli` 와 같은 `jcli` 바이너리를 설치하므로, 설치되어 있다면 먼저 `brew uninstall jenkins-cli` 로 제거합니다.

## Get started

Read the [official document](http://jcli.jenkins-zh.cn/) for more details on how to use `jcli`.

Or, you can take [a live interactive course](https://www.katacoda.com/jenkins-zh/scenarios/course-jcli) of Jenkins CLI.

## Plugins

Jenkins CLI could have more features by installing a plugin for it. You can install a plugin by the following command:

```text
jcli config plugin fetch
jcli config plugin install account
```

All official plugins could be found at [here](https://github.com/jenkins-zh/jcli-plugins).

## Contribution

If you're interested in this project. Please go through the [contribution guide](https://github.com/jenkins-zh/jenkins-cli/tree/cb3d358df4699db11b681eb0ab9adffbfb8a7bd4/CONTRIBUTING.md). Any contributions are welcome.

Thanks to JetBrains for giving us the open source licence.  
[![goland.svg](docs/book/.gitbook/assets/goland.svg)](https://www.jetbrains.com/?from=jenkins-cli)

## Similar Projects

There're a few similar projects that you might be interested in:

* [jenni](https://github.com/m-sureshraj/jenni) is a Jenkins Personal Assistant
* [jenkins-job-cli](https://github.com/gocruncher/jenkins-job-cli) Easy way to run Jenkins job from the Command Line

## Stargazers over time

[![Star History Chart](https://api.star-history.com/svg?repos=jenkins-zh/jenkins-cli&amp;type=Date)](https://star-history.com/#jenkins-zh/jenkins-cli&amp;Date)

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjenkins-zh%2Fjenkins-cli.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjenkins-zh%2Fjenkins-cli?ref=badge_large)
