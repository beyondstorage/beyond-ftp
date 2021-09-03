package utils

import (
	"bytes"
	"fmt"
	"io"
	"sync/atomic"

	_ "github.com/beyondstorage/go-service-memory"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/beyondstorage/go-stream"
)

const (
	upperStorageConnString = "memory://"
)

var (
	s        *stream.Stream
	branchId uint64
)

func NewStoragerFromString(connString string) (types.Storager, error) {
	return services.NewStoragerFromString(connString)
}

type StoragerWriter struct {
	b *stream.Branch

	path     string
	storager types.Storager
}

func (x *StoragerWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if x.b != nil {
		return x.b.ReadFrom(r)
	}

	file := new(bytes.Buffer)
	size, err := io.Copy(file, r)
	if err != nil {
		return 0, err
	}

	return x.storager.Write(x.path, file, size)
}

func (x *StoragerWriter) Complete() error {
	if x.b != nil {
		return x.b.Complete()
	}
	return nil
}

func NewStoragerWriter(path string, storager types.Storager) *StoragerWriter {
	if s != nil {
		b, err := s.StartBranch(atomic.AddUint64(&branchId, 1), path)
		if err == nil {
			return &StoragerWriter{b: b, storager: storager}
		}
	}

	return &StoragerWriter{path: path, storager: storager}
}

func StartStream(under types.Storager) {
	var err error
	s, err = newStream(stream.PersistMethodMultipart, under)
	if err == nil {
		s.Serve()
	}
}

func newStream(persisMethod string, under types.Storager) (*stream.Stream, error) {
	upper, err := NewStoragerFromString(fmt.Sprintf("%s/%s", upperStorageConnString, persisMethod))
	if err != nil {
		return nil, err
	}

	return stream.NewWithConfig(&stream.Config{
		Upper:         upper,
		Under:         under,
		PersistMethod: persisMethod,
	})
}
