package minirpc

import (
	"io"
)

type Head struct {
	Method string
	Seq    uint64
	Error  string
}

type Codec interface {
	io.Closer
	ReadHead(*Head) error
	ReadBody(any) error
	Write(*Head, any) error
}

type NewCodeFunc func(io.ReadWriteCloser) Codec

const (
	GobType  string = "application/gob"
	JsonType string = "application/json"
)

var NewCodecFuncMap map[string]NewCodeFunc

func init() {
	NewCodecFuncMap = make(map[string]NewCodeFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
