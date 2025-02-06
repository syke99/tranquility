package tranquility

import "context"

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
func WithErrorHandler[In any, Out any](errorHandler func(ctx context.Context, err error) (int, error)) func(*Handler[In, Out]) {
	return func(h *Handler[In, Out]) {
		h.errorHandler = errorHandler
	}
}
