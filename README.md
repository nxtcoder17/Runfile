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
- [x] Run tasks with dotenv files as their environment variables
- [x] Running tasks in different working directory [reference](https://taskfile.dev/reference/schema/#task)
- [x] Running tasks in parallel
- [ ] Running tasks with watch mode
- [x] Requirements prior to running a target
- [x] Environment validations and default value

### Example

```yaml
version: 0.0.1

tasks:
  test:
    env:
      key1: value1
      key2: value2
      key3:
        sh: echo -n "hello"
      dotenv:
        - .secrets/env # load dotenv file
    cmd:
      - echo "value of key1 is '$key1'"
      - echo "value of key2 is '$key2'"
      - echo "value of key3 is '$key3'"
      - echo "value of key4 is '$key4'" # assuming key4 is defined in .secrets/env
```

## Updates with example runfile with all the features

```yaml
version: 0.0.1

tasks:
  test:
    env:
      key1: value1
      key2: value2
      key3:
        sh: echo -n "hello"
      key4:
        required: true
      dotenv:
        - .secrets/env # load dotenv file
    cmd:
      - echo "value of key1 is '$key1'"
      - echo "value of key2 is '$key2'"
      - echo "value of key3 is '$key3'"
      - echo "value of key4 is '$key4'" # assuming key4 is defined in .secrets/env
  build:
    dir: cmd/app
    cmd:
      - go build -o app
  run:
    dir: cmd/app
    cmd:
      - go run .
```
