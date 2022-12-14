package main

import (
	"fmt"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/encoding/json"
	//"encoding/json"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/test/example"
)

func main() {
	bol := false
	str := `st"ring`
	var in32 int32 = 32
	var in64 int64 = 64
	var uin32 uint32 = 32
	var uin64 uint64 = 64
	var flt32 float32 = 32.123
	var flt64 float64 = 64.123
	byts := []byte("bytes")
	msg := &example.Msg{
		Str: str,
	}

	fmt.Println("测试非空对象：")
	e1 := &example.Example{
		Bol:       bol,
		Str:       str,
		In32:      in32,
		In64:      in64,
		Uin32:     uin32,
		Uin64:     uin64,
		Flt32:     flt32,
		Flt64:     flt64,
		Byts:      byts,
		Typ:       example.Typ(2),
		NestedTyp: example.Example_NestedTyp(2),
		Msg:       msg,
		NestedMsg: &example.Example_NestedMsg{
			Str: "string",
		},
		StrArr:    []string{"string"},
		TypArr:    []example.Typ{example.Typ(1), example.Typ(2)},
		StringMap: map[string]string{"key1": "val1", "key2": "val2"},
		TestOneof: &example.Example_OneofMsg{
			OneofMsg: msg,
		},
	}
	ret1, err := json.Marshal(e1)
	fmt.Println(string(ret1))
	fmt.Println(err)

	var e2 *example.Example = &example.Example{}
	json.Unmarshal(ret1, e2)
	fmt.Println(e2)
}
