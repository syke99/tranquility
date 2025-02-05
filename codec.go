package tranquility

// Codec interface can be used to provide custom
// serialization of an In and an Out so that
// you don't have to rely on tranquility's default
// json serialization format
type Codec[In any, Out any] interface {
	Marshal(out *Out) ([]byte, error)
	Unmarshal(data []byte, in *In) error
}
