#########################################################################
# @file build.sh
# @author jsf
# @mail superjsf2010@126.com
# @date 2022-07-08 22:00:19
#########################################################################
#!/bin/bash
rm -rf output
mkdir -p output/bin
go build -o ./output/bin/protoc-gen-fastjsonpb
