package bash

type Marshaller interface {
	Marshal(ts int64, command string) ([]byte, error)
	Unmarshal(b []byte) (ts int64, command string, err error)
}

var _ Marshaller = (*bashMarshaller)(nil)

type bashMarshaller struct{}

func (bashMarshaller) Marshal(ts int64, command string) ([]byte, error) {
	return []byte(command), nil
}

func (bashMarshaller) Unmarshal(b []byte) (ts int64, command string, err error) {
	return 0, string(b), nil
}
