package json

import (
	"github.com/superjsf2010/protoc-gen-fastjsonpb/x/buffer"
	"github.com/superjsf2010/protoc-gen-fastjsonpb/x/jsonparser"
)

type FastJsonpb interface {
	FastMarshal(buf *buffer.Buffer)
	FastUnmarshal(p *jsonparser.Parser)
}
