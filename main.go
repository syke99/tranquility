package tranquility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Codec interface can be used to provide custom
// serialization of an In and an Out so that
// you don't have to rely on tranquility's default
// json serialization format
type Codec[In any, Out any] interface {
	Marshal(out *Out) ([]byte, error)
	Unmarshal(data []byte, in *In) error
}

// Handler groups a generic handler func with any func for custom headers,
// serialization(both marshalling and unmarshalling), and custom error handling
// added. The structure of the incoming request body gets unmarshalled to In,
// and Out will get marshalled to the response body. Because of this, the default
// method for marshalling and unmarshalling using tranquility is via json. However,
// a Codec may be provided to implement custom serialization. If you need access to
// the entire incoming request, you can find it in the injected context using the
// "request" key
type Handler[In any, Out any] struct {
	handler       func(ctx context.Context, in *In) (*Out, error)
	headerFunc    func(ctx context.Context, in *In, out *Out) map[string]string
	codec         Codec[In, Out]
	marshallFunc  func(out *Out) ([]byte, error)
	unmarshalFunc func(data []byte, in *In) error
	errorHandler  func(err error) (int, error)
}

// WithCodec allows you to provide a codec for your tranquility
// handler to be able to inject custom serialization of your
// incoming request body and outgoing response body
func WithCodec[In any, Out any](codec Codec[In, Out]) func(*Handler[In, Out]) {
	return func(h *Handler[In, Out]) {
		h.codec = codec
	}
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
	ctx := context.WithValue(context.Background(), "request", r)

	in := new(In)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	if h.codec != nil {
		err = h.codec.Unmarshal(body, in)
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

	if h.codec != nil {
		resultBytes, err = h.codec.Marshal(out)
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
