# 🌐 endy
Simple and fast library to integrate end-to-end tests to your Golang app

# Example

- before executing code create .yaml file with description of expected tests to be executed, see **examples/** directory
- tester ends successfully, if all tests were done, otherwise, it wil os.Exit(1)
- this code showed very useful for apps, that must be tested in Gitlab CI or another CI Software, try!
- choose, what you expect to do if your tests failed, i recommend to user os.Exit(1)

```go
import (
"github.com/gonnafaraway/endy"
"time"
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
# Default
| Timeout    | 
|:-----------| 
| 10 seconds |