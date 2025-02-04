package tranquility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// TODO: doc comments

// Handler groups a generic handler func with any func for custom headers
// or error handling added. The structure of the incoming request body
// gets unmarshalled to In, and Out will get marshalled to the response body;
// because of this, tranquility relies on json. Additional information about
// the request such as any headers, the url, the method used, etc can be accessed
// through the injected Context
type Handler[In any, Out any] struct {
	handler      func(ctx context.Context, in *In) (*Out, error)
	headerFunc   func(ctx context.Context, in *In, out *Out) map[string]string
	errorHandler func(err error) (int, error)
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
	// TODO: inject rest of info about request to ctx
	ctx := r.Context()

	in := new(In)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, in)
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

	res, err := json.Marshal(out)
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

	_, _ = w.Write(res)
}
