package main

import (
	"encoding/json"
	"fmt"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/x/jsonparser"
)

func main() {
	str := `{"str":"te\"st","float":-10.123,"nested":{"str":"test"},"arr":[10,20],"bol":true,"bol1":false,"empty_obj":{},"empty_arr":[],"null":null}`
	p := jsonparser.New([]byte(str))
	m := p.Parse()
	fmt.Println(m)
	ret, _ := json.Marshal(m)
	fmt.Println(string(ret))
}
