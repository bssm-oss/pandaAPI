# pandaapi

`pandaapi`는 ChatGPT와 Google Gemini 웹 화면을 브라우저 자동화로 다루는 Go CLI입니다. 공식 API를 쓰는 대신 실제 웹 UI에 접속해서 로그인, 쿠키 저장, 질문 전송, 응답 추출을 수행합니다.

가장 쉽게 설치하는 방법은 아래 한 줄입니다.

```bash
go install github.com/bssm-oss/pandaAPI@latest
```

설치가 끝나면 `pandaapi` 바이너리가 Go bin 경로에 생깁니다.

## 이게 뭐하는 도구인가요?

이 프로젝트는 이런 용도로 씁니다.

- `pandaapi auth --provider ...`로 브라우저 로그인 세션을 저장
- `pandaapi ask --query ...`로 ChatGPT 또는 Gemini에 질문 전송
- ChatGPT는 저장된 로그인 쿠키를 기반으로 동작
- Gemini는 로그인 쿠키가 있으면 인증 세션을 사용하고, 쿠키가 없으면 현재 공개된 signed-out Gemini 웹 흐름으로 질문을 시도

즉, “브라우저에서 직접 쓰는 AI 웹앱을 CLI로 자동화”하는 도구입니다.

## 빠른 설치

### 1) GitHub에서 바로 설치

가장 쉬운 설치 방법입니다.

```bash
go install github.com/bssm-oss/pandaAPI@latest
```

### 2) 저장소를 받아 직접 빌드

```bash
make deps
make build
```

## 실행에 필요한 것

- Go 1.21+
- `ask`용 Lightpanda CDP 서버
  - 기본 주소: `http://127.0.0.1:9222`
- `auth`용 로컬 Chrome 또는 Chromium

## 환경 변수

- `LIGHTPANDA_CDP`: Lightpanda CDP 주소 변경
- `PANDAAPI_COOKIE_DIR`: 쿠키 저장 디렉터리 변경

## 가장 쉬운 사용 순서

### ChatGPT

1. 로그인 세션 저장

```bash
pandaapi auth --provider chatgpt
```

2. 질문 보내기

```bash
pandaapi ask --query "Hello" --provider chatgpt
```

### Gemini

#### 방법 1: 로그인해서 사용

```bash
pandaapi auth --provider gemini
pandaapi ask --query "Hello" --provider gemini
```

#### 방법 2: 로그인 없이 사용 시도

현재 Gemini는 쿠키가 없을 때 로컬 Chrome 런타임으로 fallback해서 signed-out 웹 흐름을 시도합니다.

```bash
PANDAAPI_COOKIE_DIR=$(mktemp -d) pandaapi ask --query "Say exactly: PONG" --provider gemini
```

`--provider`를 생략하면 기본값은 `chatgpt`입니다.

## 명령어 요약

### 인증 세션 저장

```bash
pandaapi auth --provider chatgpt
pandaapi auth --provider gemini
```

### 질문 보내기

```bash
pandaapi ask --query "Hello" --provider chatgpt
pandaapi ask --query "Summarize this page" --provider gemini
```

## 현재 동작 방식

- `auth`
  - 보이는 로컬 브라우저를 열어 수동 로그인 후 쿠키를 저장합니다.
- `ask`
  - ChatGPT: 저장된 쿠키가 필요합니다.
  - Gemini: 저장된 쿠키가 있으면 그 세션을 쓰고, 없으면 로컬 Chrome으로 signed-out 질문 경로를 시도합니다.
- 쿠키 기본 저장 경로는 `~/.pandaapi`입니다.

## 프로젝트 구조

```text
pandaapi/
├── .github/workflows/ci.yml
├── AGENTS.md
├── docs/changes/2026-04-06-lightpanda-cli.md
├── go.mod
├── go.sum
├── main.go
├── auth.go
├── ask.go
├── browser.go
├── chatgpt.go
├── gemini.go
├── cookies.go
├── config.go
├── config_test.go
├── main_test.go
├── Makefile
└── README.md
```

## 검증 방법

```bash
make deps
go vet ./...
make test
make build
GOBIN=$PWD/.bin go install .
```

수동 확인 예시는 아래와 같습니다.

- `./pandaapi`
- `./pandaapi auth --provider chatgpt`
- `./pandaapi auth --provider gemini`
- `./pandaapi ask --query "Hello" --provider chatgpt`
- `./pandaapi ask --query "Hello" --provider gemini`
- `PANDAAPI_COOKIE_DIR=$(mktemp -d) ./pandaapi ask --query "Say exactly: PONG" --provider gemini`

## 제한 사항

- ChatGPT와 Gemini는 웹 UI 구조가 자주 바뀔 수 있어서 selector 유지보수가 필요합니다.
- `ask`는 외부 웹 서비스와 브라우저 런타임에 의존합니다.
- Lightpanda CDP가 떠 있지 않으면 Lightpanda 경로 검증은 진행할 수 없습니다.
- ChatGPT는 현재 확인된 환경에서 Lightpanda가 브라우저 검증 페이지에 걸릴 수 있습니다.
- Gemini 로그인 없는 fallback은 현재 공개된 Gemini signed-out 웹 UI가 입력창을 계속 제공한다는 전제에 의존합니다.

## 개발 명령

- `make deps`: 모듈 의존성 정리
- `make test`: 테스트 실행
- `make build`: 바이너리 빌드
- `make clean`: 로컬 바이너리 삭제

## GitHub 설치 참고

- 배포 경로는 `github.com/bssm-oss/pandaAPI`입니다.
- `go install github.com/bssm-oss/pandaAPI@latest`로 설치할 수 있습니다.
- 설치 후에도 런타임 요구사항은 그대로입니다.
  - ChatGPT/Gemini 웹 접근 가능
  - Lightpanda 또는 로컬 Chrome/Chromium 사용 가능
