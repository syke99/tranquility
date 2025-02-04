package tranquility

import (
	"context"
	"errors"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("GET /hello", NewHandler(
		HelloWorldHandler,
		WithErrorHandler[Fizz, Buzz](ErrorHandler),
	))
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

func ErrorHandler(err error) (int, error) {
	if errors.Is(BadLanguage, err) {
		// do any custom error handling based on the specific types of errors and
		// return the appropriate status code, and the newly handled error
		return http.StatusBadRequest, err
	}
	return http.StatusInternalServerError, err
}
