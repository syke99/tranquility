package tranquility_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syke99/tranquility"
)

var BadLanguage = errors.New("language not supported")

type MyCodec[In any, Out any] struct{}

var TestCodec = &MyCodec[Fizz, Buzz]{}

func (c *MyCodec[In, Out]) Marshal(out *Out) ([]byte, error) {
	return json.Marshal(out)
}

func (c *MyCodec[In, Out]) Unmarshal(data []byte, in *In) error {
	return json.Unmarshal(data, in)
}

type Fizz struct {
	Language string `json:"language"`
}

type Buzz struct {
	Greeting string `json:"greeting"`
}

var TestHandler = func(ctx context.Context, in *Fizz) (*Buzz, error) {
	if in.Language != "english" {
		return nil, BadLanguage
	}
	return &Buzz{
		Greeting: "hello world!",
	}, nil
}

var TestHeaderFunc = func(ctx context.Context, in *Fizz, out *Buzz) map[string]string {
	return map[string]string{
		"x-language":   in.Language,
		"Content-Type": "application/json",
	}
}

var TestErrorHandler = func(ctx context.Context, err error) (int, error) {
	if errors.Is(BadLanguage, err) {
		// do any custom error handling based on the specific types of errors and
		// return the appropriate status code, and the newly handled error
		return http.StatusBadRequest, err
	}
	return http.StatusInternalServerError, err
}

func TestNewHandler(t *testing.T) {
	handler := tranquility.NewHandler(TestHandler)

	assert.NotNil(t, handler)
}

func TestHandlerWithOptions(t *testing.T) {
	handler := tranquility.NewHandler(
		TestHandler,
		tranquility.WithHeaderFunc(TestHeaderFunc),
		tranquility.WithCodec[Fizz, Buzz](TestCodec),
		tranquility.WithErrorHandler[Fizz, Buzz](TestErrorHandler),
	)

	assert.NotNil(t, handler)
}

func TestHandlerSuccess(t *testing.T) {
	mux := http.NewServeMux()

	handler := tranquility.NewHandler(
		TestHandler,
		tranquility.WithHeaderFunc(TestHeaderFunc),
		tranquility.WithCodec[Fizz, Buzz](TestCodec),
		tranquility.WithErrorHandler[Fizz, Buzz](TestErrorHandler),
	)

	mux.Handle("GET /hello", handler)

	fizz := &Fizz{Language: "english"}

	b, err := json.Marshal(fizz)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/hello", bytes.NewReader(b))

	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	//languageHeader := res.Header.Get("x-language")
	//
	//assert.Equal(t, "english", languageHeader)

	buzz := &Buzz{Greeting: "hello world!"}

	b, err = json.Marshal(buzz)
	assert.NoError(t, err)

	resBytes, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, string(b), string(resBytes))
}

func TestHandlerError(t *testing.T) {
	mux := http.NewServeMux()

	handler := tranquility.NewHandler(
		TestHandler,
		tranquility.WithHeaderFunc(TestHeaderFunc),
		tranquility.WithCodec[Fizz, Buzz](TestCodec),
		tranquility.WithErrorHandler[Fizz, Buzz](TestErrorHandler),
	)

	mux.Handle("GET /hello", handler)

	fizz := &Fizz{Language: "latin"}

	b, err := json.Marshal(fizz)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/hello", bytes.NewReader(b))

	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	resBytes, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, BadLanguage.Error(), strings.TrimSuffix(string(resBytes), "\n"))
}

func TestHandlerHeaders(t *testing.T) {
	mux := http.NewServeMux()

	handler := tranquility.NewHandler(
		TestHandler,
		tranquility.WithHeaderFunc(TestHeaderFunc),
		tranquility.WithCodec[Fizz, Buzz](TestCodec),
		tranquility.WithErrorHandler[Fizz, Buzz](TestErrorHandler),
	)

	mux.Handle("GET /hello", handler)

	fizz := &Fizz{Language: "english"}

	b, err := json.Marshal(fizz)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/hello", bytes.NewReader(b))

	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, "english", res.Header.Get("x-language"))
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}
