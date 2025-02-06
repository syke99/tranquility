package tranquility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Handler groups a generic handler func with any func for custom headers,
// serialization(both marshalling and unmarshalling), and custom error handling
// added. The structure of the incoming request body gets unmarshalled to In,
// and Out will get marshalled to the response body. Because of this, the default
// method for marshalling and unmarshalling using tranquility is via json. However,
// a Codec may be provided to implement custom serialization. If you need access to
// the entire incoming request, you can find it in the injected context using the
// "request" key
type Handler[In any, Out any] struct {
	handler      func(ctx context.Context, in *In) (*Out, error)
	headerFunc   func(ctx context.Context, in *In, out *Out) map[string]string
	codec        Codec[In, Out]
	errorHandler func(ctx context.Context, err error) (int, error)
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

	ctx.Value("request")

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

			resCode, resErr = h.errorHandler(ctx, err)
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
