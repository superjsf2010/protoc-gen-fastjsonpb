package main

import (
	"testing"

	json "encoding/json"

	jsoniter "github.com/json-iterator/go"
	fastjsonpb "github.com/superjsf2010/protoc-gen-fastjsonpb/encoding/json"
	"github.com/superjsf2010/protoc-gen-fastjsonpb/test/example"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

var byts []byte = []byte(`{"str":"string","in32":32,"in64":64,"uin32":32,"uin64":64,"flt32":32.123,"flt64":64.123,"byts":"Ynl0ZXM=","typ":"TYPB","msg":{"str":"string"},"strArr":["string"],"typArr":["TYPA","TYPB"],"stringMap":{"key1":"val1","key2":"val2"},"nestedTyp":"TYPB","nestedMsg":{"str":"string"},"oneofBol":false,"unknown":{"str":"string","in32":32,"in64":64,"uin32":32,"uin64":64,"flt32":32.123,"flt64":64.123,"byts":"Ynl0ZXM="}}`)

func BenchmarkFastJsonpbUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := example.ExampleNew()
		fastjsonpb.Unmarshal(byts, e)
		// 对象析构入Pool
		e.Destructor()
	}
}

func BenchmarkStdJsonpbUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := &example.Example{}
		jsonpb.Unmarshal(byts, e)
	}
}

func BenchmarkStdJsonUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := &example.Example{}
		json.Unmarshal(byts, e)
	}
}

func BenchmarkJsoniterUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := &example.Example{}
		jsoniter.Unmarshal(byts, e)
	}
}
