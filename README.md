## Runfile is a super simple task runner

All it does is run some pre-configured tasks for you, like running your applications, tests, building binaries, or some other scripts.

It is inspired by other task runners like [Taskfile](https://taskfile.dev), [Make](https://www.gnu.org/software/make/manual/make.html) etc.
I decided to build my version of it, accounting the experiences that i want to see in a task runner. 

### Features

- [x] Run tasks (commands)
- [x] Run tasks with Key-Value environment variables
- [x] Run tasks with dynamic environment variables (referencing values by shell execution)
- [x] Run tasks with dotenv files as their environment variables
- [x] Importing tasks from different working directory (must in a monorepo) [reference](https://taskfile.dev/reference/schema/#task)
- [x] Running tasks in parallel (e.g. css generation and build in parallel)
- [x] Running tasks with watch mode (e.g. like hot reload)
- [x] Requirements prior to running a target (e.g. sanity tests)
- [x] Environment validations and default value

### Installation

| Tool  | Command                                                   |
| :---: | :---:                                                     |
| Go    | `go install github.com/nxtcoder17/runfile/cmd/run@latest` |


```bash
go install github.com/nxtcoder17/runfile/cmd/run@latest
```

## Usage
    
### Runfile

Create a `Runfile` in the root of your project, and add tasks to it.

### Examples

1. simple tasks

```yaml
tasks:
  example:
    cmd:
        - echo "example"
```

![Output](https://github.com/user-attachments/assets/37e33c94-bd3d-407a-8cbe-e62f6e3e2411)

2. using environment variables

```yaml
tasks:
  example:
    env:
      key: "hello world"
    cmd:
      - echo $key
```

![Output](https://github.com/user-attachments/assets/3a0979e6-53a4-4979-bc93-de6adc36fe74)

3. using dynamic environment variables

```yaml
tasks:
  example:
    env:
      key: 
        sh: echo $HOME
    cmd:
      - echo $key
```

![Image](https://github.com/user-attachments/assets/f9ca3ae6-4a49-46e4-a07b-66ba84ba14a3)

4. using dotenv based environment variables

```yaml
tasks:
  example:
    dotenv:
      - .env
    cmd:
      - echo $key
```

```bash
# file: .env
key="my-dotenv-secret"
```

![Image](https://github.com/user-attachments/assets/941b6a9d-57ae-46f1-a320-e76278d6b1e2)

5. validating required environment variable

```yaml
tasks:
  example:
    env:
      key:
        required: true
    cmd:
      - echo $key
```
![Image](https://github.com/user-attachments/assets/b900d81a-0eae-4ef6-ac30-8b50cf4ce292)

6. referencing other tasks

```yaml
tasks:
  script1:
    cmd:
      - echo "i am script 1 (key=$key)"

  example:
    cmd:
      - run: script1
        env:
          key: "hello"
      - echo this is example (key=$key)
```
![Image](https://github.com/user-attachments/assets/8e7c2d4b-b9e1-4e0e-9b07-e012adc86d64)
