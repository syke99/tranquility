package tranquility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Handler groups a generic handler func with any func for custom headers
// or error handling added. The structure of the incoming request body
// gets unmarshalled to In, and Out will get marshalled to the response body;
// because of this, the default method for marshalling and unmarshalling using
// tranquility is via json. However, a custom (Un)MarshallFunc can be provided
// using the WithCustom(Un)MarshallFunc option(s) whenever creating a new handler
// with tranquility to allow you to use any type of serialization. If you need
// access to the entire incoming request, you can find it in the injected context
// using the "request" key
type Handler[In any, Out any] struct {
	handler       func(ctx context.Context, in *In) (*Out, error)
	headerFunc    func(ctx context.Context, in *In, out *Out) map[string]string
	marshallFunc  func() ([]byte, error)
	unmarshalFunc func(data []byte) (int, error)
	errorHandler  func(err error) (int, error)
}

// WithHeaderFunc allows you to define any custom headers to
// be added to a successful request before the response is written back
func WithHeaderFunc[In any, Out any](headerFunc func(ctx context.Context, in *In, out *Out) map[string]string) func(*Handler[In, Out]) {
	return func(h *Handler[In, Out]) {
		h.headerFunc = headerFunc
	}
}

// WithErrorHandler allows you to inject custom error handling
// into your tranquility Handler. This is where you define the specific
// status code to be returned with an error, along with any error
// manipulation you may want to perform
func WithErrorHandler[In any, Out any](errorHandler func(err error) (int, error)) func(*Handler[In, Out]) {
	return func(h *Handler[In, Out]) {
		h.errorHandler = errorHandler
	}
}

// WithCustomUnmarshallFunc allows you to provide a custom method
// for unmarshalling instead of letting tranquility default to json
func WithCustomUnmarshallFunc[In any, Out any](unmarshallFunc func(data []byte) (int, error)) func(*Handler[In, Out]) {
	return func(h *Handler[In, Out]) {
		h.unmarshalFunc = unmarshallFunc
	}
}

// WithCustomMarshallFunc allows you to provide a custom method
// for marshalling instead of letting tranquility default to json
func WithCustomMarshallFunc[In any, Out any](marshallFunc func() ([]byte, error)) func(*Handler[In, Out]) {
	return func(h *Handler[In, Out]) {
		h.marshallFunc = marshallFunc
	}
}

func NewHandler[In any, Out any](handler func(ctx context.Context, in *In) (*Out, error), opts ...func(Handler *Handler[In, Out])) http.Handler {
	h := &Handler[In, Out]{
		handler: handler,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *Handler[In, Out]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(context.Background(), "request", r)

	in := new(In)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	if h.unmarshalFunc != nil {
		_, err = h.unmarshalFunc(body)
	} else {
		err = json.Unmarshal(body, in)
	}

	if err != nil {
		http.Error(w, "unable to unmarshal request body", http.StatusBadRequest)
		return
	}

	out, err := h.handler(ctx, in)
	if err != nil {
		if h.errorHandler != nil {
			resCode := http.StatusInternalServerError
			resErr := err

			resCode, resErr = h.errorHandler(err)
			http.Error(w, resErr.Error(), resCode)
			return
		}
	}

	var resultBytes []byte

	if h.marshallFunc != nil {
		resultBytes, err = h.marshallFunc()
	} else {
		resultBytes, err = json.Marshal(out)
	}

	if err != nil {
		http.Error(w, "unable to marshal response", http.StatusInternalServerError)
		return
	}

	if h.headerFunc != nil {
		customHeaders := h.headerFunc(ctx, in, out)

		for k, v := range customHeaders {
			w.Header().Set(k, v)
		}
	}

	_, _ = w.Write(resultBytes)
}
