package utils

import (
	"fmt"
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
	streamMap    map[string]*stream.Stream
	branchId     uint64
)

func NewStoragerFromString(connString string) (types.Storager, error) {
	return services.NewStoragerFromString(connString)
}

func Branch(label, path string) *stream.Branch {
	if s, ok := streamMap[label]; ok {
		b, err := s.StartBranch(atomic.AddUint64(&branchId, 1), path)
		if err == nil {
			return b
		}
	}
	return nil
}

func StartStream(under types.Storager) {
	streamMap = make(map[string]*stream.Stream)
	for _, method := range []string{stream.PersistMethodAppend, stream.PersistMethodMultipart} {
		s, err := newStream(method, under)
		if err == nil {
			go s.Serve()
			streamMap[method] = s
		}
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
