package bash

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

var _ Marshaller = (*marshaller)(nil)

func NewMarshaller() Marshaller {
	return marshaller{}
}

type marshaller struct{}

func (marshaller) Marshal(ts int64, command string) ([]byte, error) {
	return []byte(command), nil
}

func (marshaller) Unmarshal(b []byte) (ts int64, command string, err error) {
	return 0, string(b), nil
}
