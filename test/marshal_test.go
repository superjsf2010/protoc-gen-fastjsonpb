package main

import (
	"testing"

	json "encoding/json"

	jsoniter "github.com/json-iterator/go"
	fastjsonpb "github.com/superjsf2010/protoc-gen-fastjsonpb/encoding/json"
	"github.com/superjsf2010/protoc-gen-fastjsonpb/test/example"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

var msg *example.Msg = &example.Msg{
	Bol:   true,
	Str:   "st\"ring",
	In32:  32,
	In64:  64,
	Uin32: 32,
	Uin64: 64,
	Flt32: 32.123,
	Flt64: 64.124,
	Byts:  []byte("bytes"),
}

var e *example.Example = &example.Example{
	Bol:       true,
	Str:       "s\"tring",
	In32:      32,
	In64:      64,
	Uin32:     32,
	Uin64:     64,
	Flt32:     32.123,
	Flt64:     64.124,
	Byts:      []byte("bytes"),
	Typ:       example.Typ(2),
	NestedTyp: example.Example_NestedTyp(2),
	Msg:       msg,
	NestedMsg: &example.Example_NestedMsg{
		Str: "string",
	},
	StrArr: []string{"string"},
	TypArr: []example.Typ{example.Typ(1), example.Typ(2)},
	TestOneof: &example.Example_OneofBol{
		OneofBol: false,
	},
}

func BenchmarkFastJsonpbMarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fastjsonpb.Marshal(e)
	}
}

func BenchmarkStdJsonpbMarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonpb.Marshal(e)
	}
}

func BenchmarkStdJsonMarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(e)
	}
}

func BenchmarkJsoniterMarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsoniter.Marshal(e)
	}
}
