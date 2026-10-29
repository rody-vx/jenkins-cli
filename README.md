# Jenkins CLI (jcli)

[jenkins-zh/jenkins-cli](https://github.com/jenkins-zh/jenkins-cli) 의 fork 입니다.
여러 Jenkins 서버를 등록해 두고 job 검색 · 빌드 · 로그 확인 등을 명령행에서 처리합니다.

## 원본 대비 변경

- `job search` 를 플러그인 없이 사용
  - Jenkins 기본 JSON API 로 폴더 안까지 검색한다 (원본은 `pipeline-restful-api` 플러그인 필요)
  - 서버 설정의 `search` 로 방식 선택: `api`(기본) · `plugin`. `-p` 옵션을 주면 플러그인 사용
- job 이름을 `폴더/잡` 형식으로 지정 (`job/폴더/job/잡`, `폴더 잡` 형식도 그대로 지원)
- `jcli version` · `jcli version upgrade` 가 이 저장소의 Release 를 기준으로 동작

## 설치

[Homebrew](https://brew.sh/) (macOS, Linux) 로만 배포합니다.

```shell
brew install rody-vx/tap/jcli
```

| 업그레이드 | 삭제 |
|---|---|
| `brew upgrade jcli` | `brew uninstall jcli` |

공식 `jenkins-cli` 와 같은 `jcli` 바이너리를 설치하므로, 설치되어 있다면 먼저 `brew uninstall jenkins-cli` 로 제거합니다.

## 설정

Jenkins 사용자 설정에서 API Token 을 발급한 뒤 서버를 등록합니다. 설정은 `~/.jenkins-cli.yaml` 에 저장됩니다.

```shell
jcli config add -n qa --url https://jenkins.example.com -u <사용자> -t <API 토큰>
jcli config list
jcli config select          # 기본 서버 변경
jcli -j qa job history ...  # 이번 명령만 다른 서버로
```

```yaml
current: qa
jenkins_servers:
- name: qa
  url: https://jenkins.example.com
  username: <사용자>
  token: <API 토큰>
  search: api   # 생략하면 api
```

토큰이 평문으로 저장되므로 `chmod 600 ~/.jenkins-cli.yaml` 을 권장합니다.

## 사용 예

```shell
jcli job search deploy                    # 이름에 deploy 가 들어간 job
jcli job search --type Folder             # 폴더만
jcli job param 폴더/잡                     # 파라미터 확인
jcli job build 폴더/잡 -b -l               # 확인 없이 빌드하고 로그 따라가기
jcli job build 폴더/잡 -b --param-entry BRANCH=main
jcli job history 폴더/잡
jcli job log 폴더/잡 -w -e                 # 로그를 따라가다 실패하면 exit 1
jcli job stop 폴더/잡 <빌드번호> -b
```

검색 결과의 `URL` 값(`job/폴더/job/잡/`)도 job 이름으로 그대로 쓸 수 있습니다.

## 릴리즈

`v*` 태그를 push 하면 GitHub Actions 가 macOS · Linux(arm64, amd64) 바이너리를 빌드해 Release 에 올립니다.
이후 [rody-vx/homebrew-tap](https://github.com/rody-vx/homebrew-tap) 의 `Formula/jcli.rb` 에서 `version` · `url` · `sha256` 을 갱신합니다.

## License

[MIT](LICENSE)
