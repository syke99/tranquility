# tranquility

[![Go Reference](https://pkg.go.dev/badge/github.com/syke99/tranquility.svg)](https://pkg.go.dev/github.com/syke99/tranquility)
[![Go Reportcard](https://goreportcard.com/badge/github.com/syke99/gworker)](https://goreportcard.com/report/github.com/syke99/tranquility)

[![LICENSE](https://img.shields.io/github/license/syke99/tranquility)](https://pkg.go.dev/github.com/syke99/tranquility/blob/master/LICENSE)

easily write your HTTP handlers in Go with Generics

Why tranquility?
====

`tranquility` allows you to write HTTP handlers where you can easily visualize the structure of the incoming request,
as well as the outgoing response, by leveraging Generics and still being `net/http` compliant.
This is inspired by Go's design philosophy of focusing on being readable. On top of that, if you so choose, you can structure
your code to keep a handler and its models together. Eg:

```go
-- handlers
 |
 |__ fizz
 | |__ buzz.go // model for incoming request
 | |__ bazz.go // model for outgoing response
 | |__ handler.go // fizz handler (contains business logic)
 |
 |__ hello
 | |__ language.go // model for incoming request
 | |__ greeting.go // model for outgoing response
 | |__ handler.go // hello handler (contains business logic)
 |
 |...
```

How do I use tranquility?
====

### Installing
To install `tranquility` in a repo, simply run
```bash
go get github.com/syke99/tranquility
```

Then you can import the package in any go file you'd like

```go
import "github.com/syke99/tranquility"
```

### Basic Usage

First, define your models and your handler func:
```go
type Fizz struct {
	Language string `json:"language"`
}


type Buzz struct {
    Greeting string `json:"greeting"`
}

func MyHandler(ctx context.Context, in *Fizz) (*Buzz, error) {
    if in.Language != "english" {
        return nil, BadLanguage
    }
    return &Buzz{
        Greeting: "hello world!",
    }, nil
}
```

After that, simply create it as a tranquility handler:
```go
myCoolNewHandler := tranquility.NewHandler(MyHandler)
```

Then you can register this handler just like you would any other!
```go
mux := http.NewServeMux()

mux.Handle("GET /hello", myCoolNewHandler)
```

### Advanced Usage

##### Custom Serialization

`tranquility` defaults to using JSON for serializing the request and response bodies. Eg:

To use a different form of serialization, you can define a struct that satisfies
the `tranquility.Codec` interface
```go
// in this example, we'll instead use github.com/golang/protobuf/proto (un)marshaling
type MyCodec[In any, Out any] struct {}

func (c *MyCodec[In, Out]) Marshal(out *Out) ([]byte, error) {
    return proto.Marshal(out)
}

func (c *MyCodec[In, Out]) Unmarshal(data []byte, in *In) error {
    return proto.Unmarshal(data, in)
}
```

Then simply use the `tranquility.WithCodec` option whenever creating your new Handler,
passing in the Codec you just created.
```go
myCoolNewHandler := tranquility.NewHandler(
	MyHandler
	tranquility.WithCodec[Fizz, Buzz](&MyCodec{})
)
```

##### Custom Headers

You can also provide a function for adding custom headers to your response by using
`tranquility.WithHeaderFunc` whenever creating your tranquility handler and passing in a function that will return a map off
key-value strings to be added to the response headers. Eg:

First, define your function for adding headers to a response
```go
func MyHeaderFunc(ctx context.Context, in *Fizz, out *Buzz) map[string]string {
    return map[string]string{
        "x-language":   in.Language,
        "Content-Type": "application/json",
    }
}
```

Then, just like custom serialization, simply use the `tranquility.WithHederFunc` option
whenever creating your tranquility handler
```go
myCoolNewHandler := tranquility.NewHandler(
	MyHandler,
	tranquility.WithHederFunc(MyHeaderFunc)
)
```

##### Custom Headers

While tranquility defaults to a status code of 500 and just simply passing the error
back to the caller, you _can_ implement your own custom error handling by using the
`tranquility.WithErrorHandler` option whenever creating your tranquility handler. Eg.

First, define your error handler func
```go
var BadLanguage = errors.New("language not supported")

func MyErrorHandler(ctx context.Context, err error) (int, error) {
    if errors.Is(BadLanguage, err) {
        // do any custom error handling based on the specific types of errors and
        // return the appropriate status code, and the newly handled error
        return http.StatusBadRequest, err
    }
    return http.StatusInternalServerError, err
}
```

And again, it's as easy as including a call to `tranquility.WithErrorHandler` whenever
creating your tranquility handler
```go
myCoolNewHandler := tranquility.NewHandler(
	MyHandler,
	tranquility.WithErrorHandler[Fizz, Buzz](MyErrorHandler)
)
```

##### Accessing entire incoming Request

You can also access the entire incoming `*http.Request` via the injected context
in the handler and any registered options. It can be found via the, you guessed it, "request" key
```go
func MyHandler(ctx context.Context, in *Fizz) (*Buzz, error) {
    req := ctx.Value("request") // as with all context values, they must be coerced to their correct type as they're stored as an `any` type
	...
}
```
