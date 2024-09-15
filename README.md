## Runfile is a super simple task runner

All it does is run some pre-configured tasks for you, like running your applications, tests, building binaries, or some other scripts.

It is inspired by other task runners like Taskfile, Make etc.

But source code of those tools are like super big, and complex. So I decided to make a simpler one.

## Installation

```bash
go install github.com/nxtcoder17/runfile/cmd/run@latest
```

## Usage

### Runfile

Create a `Runfile` in the root of your project, and add tasks to it.

### Features

- [x] Run tasks
- [x] Run tasks with Key-Value environment variables
- [x] Run tasks with dynamic environment variables (by shell execution)

### Example

```yaml
version: 0.0.1

tasks:
  build:
    cmd: 
      - go build -o bin/run ./cmd/run
  test:
    cmd: 
      - go test ./...
