package json

import (
	"errors"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/x/jsonparser"
)

func Unmarshal(data []byte, obj interface{}) error {
	fastjsonpbObj, ok := obj.(FastJsonpb)
	if !ok {
		return errors.New("object do not implements FastJsonpb")
	}
	p := jsonparser.New(data)
	fastjsonpbObj.FastUnmarshal(p)
	return nil
}
