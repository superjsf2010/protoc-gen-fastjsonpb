package json

import (
	"errors"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/x/buffer"
)

func Marshal(obj interface{}) ([]byte, error) {
	fastjsonpbObj, ok := obj.(FastJsonpb)
	if !ok {
		return nil, errors.New("object do not implements FastJsonpb")
	}
	buf := buffer.New()
	fastjsonpbObj.FastMarshal(buf)
	buffer.BufPool.Put(buf)
	return buf.Bytes(), nil
}
