package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/gen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)
	var req pluginpb.CodeGeneratorRequest
	proto.Unmarshal(input, &req)

	generator, err := gen.New(&req)
	if err != nil {
		panic(err)
	}

	stdout, err := generator.GenerateAllFiles()
	if err != nil {
		panic(err)
	}
	out, err := proto.Marshal(stdout)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, string(out))
}
