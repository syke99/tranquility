package tranquility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("GET /hello", NewHandler(
		HelloWorldHandler,
		WithHeaderFunc(HelloWorldHeaders),
		WithCodec[Fizz, Buzz](&MyCodec{}),
		WithErrorHandler[Fizz, Buzz](ErrorHandler),
	))
}

type MyCodec struct{}

func (c *MyCodec) Marshal(out *Buzz) ([]byte, error) {
	// for simplicity sake, wrapping json.Unmarshall call
	// but this is where/how you can implement a custom
	// MarshallFunc
	return json.Marshal(out)
}

func (c *MyCodec) Unmarshal(data []byte, in *Fizz) error {
	// for simplicity sake, wrapping json.Unmarshall call
	// but this is where/how you can implement a custom
	// UnmarshallFunc
	return json.Unmarshal(data, in)
}

func (c *MyCodec) UnmarshallFizz(data []byte, fizz *Fizz) error {
	// for simplicity sake, wrapping json.Unmarshall call
	// but this is where/how you can implement a custom
	// UnmarshallFunc
	return json.Unmarshal(data, fizz)
}

func (c *MyCodec) MarshallBuzz(buzz *Buzz) ([]byte, error) {
	// for simplicity sake, wrapping json.Unmarshall call
	// but this is where/how you can implement a custom
	// MarshallFunc
	return json.Marshal(buzz)
}

type Fizz struct {
	Language string `json:"language"`
}

type Buzz struct {
	Greeting string `json:"greeting"`
}

var (
	BadLanguage = errors.New("language not supported")
	// etc
)

func HelloWorldHandler(ctx context.Context, in *Fizz) (*Buzz, error) {
	if in.Language != "english" {
		return nil, errors.New("language not supported")
	}
	return &Buzz{
		Greeting: "hello world!",
	}, nil
}

func HelloWorldHeaders(ctx context.Context, in *Fizz, out *Buzz) map[string]string {
	return map[string]string{
		"x-language": in.Language,
	}
}

func ErrorHandler(err error) (int, error) {
	if errors.Is(BadLanguage, err) {
		// do any custom error handling based on the specific types of errors and
		// return the appropriate status code, and the newly handled error
		return http.StatusBadRequest, err
	}
	return http.StatusInternalServerError, err
}
