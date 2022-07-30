#########################################################################
# @file run_bechmark.sh
# @author jsf
# @mail superjsf2010@126.com
# @date 2022-07-28 14:56:51
#########################################################################
#!/bin/bash

if [ "$1" == "b" ];then
    go test -bench=. --test.benchmem
elif [ "$1" == "g" ];then
    go-torch marshal.test cpu.prof
fi
