# Go Notifier
Implementation of a Notifier library in Go, it supports middlewares and runs handlers in goroutines. The API is thread safe.

## Uses

### Base use
```go
package main

import (
    "fmt"

    "github.com/valsov/notifier"
)

func main() {
    notifier.RegisterHandler(handler)
    notifier.RegisterMiddleware(logger)
    notifier.RegisterMiddleware(sample)

    notifier.Publish("data")

    // [...] Omitted: Need to wait here to see prints
}

func handler(x string) {
    fmt.Printf("h() => %s")
}

func logger(ctx *notifier.ExecutionContext) {
    fmt.Printf("START => %s")
    ctx.Next() // Call next handler in execution chain
    fmt.Printf("END")
}

func sample(ctx *notifier.ExecutionContext) {
    fmt.Printf("sample middleware")
    ctx.Next()
}
```

Running the code above outputs (note that middlewares registration order matters):

```
START => data
sample middleware
h() => data
END
```

It is possible to register multiple handlers for the same message type. They will all be executed following the registered middleware chain.

### Supported handler types
```go
func main() {
    sample := sampleStruct{}
    
    notifier.RegisterHandler(function)
    notifier.RegisterHandler(sample.method)
    notifier.RegisterHandler(func(parameter int){
        // [...]
    })
}

func function(parameter int) {
    // [...]
}

type sampleStruct struct {}
func (s *sampleStruct) method(){
    // [...]
}
```
