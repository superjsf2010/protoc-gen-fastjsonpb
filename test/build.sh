#########################################################################
# @file build.sh
# @author jsf
# @mail superjsf2010@126.com
# @date 2022-07-12 13:55:06
#########################################################################
#!/bin/bash
HOME_DIR=`dirname $0`;
HOME_DIR=$HOME_DIR/../;
cd $HOME_DIR;
HOME_DIR=$(pwd);

set -e;
sh build.sh;
cd -;

PROTOC=protoc;

export PATH=$HOME_DIR/output/bin:$PATH

rm -rf example;

# 需要编译的proto文件，增加新的层级proto直接在后添加即可
compile_proto_list=( \
    "*.proto" \
)

for var in ${compile_proto_list[@]};
do
    $PROTOC --go_out=$HOME_DIR \
        --fastjsonpb_out=$HOME_DIR \
        --proto_path=$HOME_DIR/test \
        $HOME_DIR/test/$var
done
