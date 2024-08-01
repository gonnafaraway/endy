# 🌐 endy
Simple and fast library to integrate end-to-end tests to your Golang app

# Features

- before executing code create .yaml file with description of expected tests to be executed, see **examples/** directory
- tester ends successfully, if all tests were done, otherwise, it wil os.Exit(1)
- this code showed very useful for apps, that must be tested in Gitlab CI or another CI Software, try!
- choose, what you expect to do if your tests failed, i recommend to use os.Exit(1)
- you can set secrets in envs using EnvSecret key in yaml file, it will be parsed and overwrite the value written before, very useful for ci/cd pipelines

# Library
```go
import (
"time"

"github.com/gonnafaraway/endy"
)

func main() {
// create new tester instance
e := endy.New()

// set timeout for all requests, 10 seconds by default
e.SetTimeout(10 * time.Second)

// set path with endpoints
e.SetConfigPath("config.yaml")

// run tests safely
if err := e.Run(); err != nil {
panic(err)}
}
```

# Benchmark
- with installed bombardier(https://github.com/codesenberg/bombardier) tool endy can benchmark endpoints and print out short results

# CLI

```
go install github.com/gonnafaraway/endy/cmd/endy@latest 
```
- if path not set, will be taken config.yaml from the same pwd
- default timeout is the same, as in library - 10 seconds

# Default
| Timeout    | 
|:-----------| 
| 10 seconds |