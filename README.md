# protoc-gen-fastjsonpb

## 安装

go install github.com/superjsf2010/protoc-gen-fastjsonpb

## 使用

### 依赖

- protoc
- protoc-gen-go

### 编译

```
$PROTOC --go_out=$OUTPUT_DIR \
      --fastjsonpb_out=$OUTPUT_DIR \
      --proto_path=$PROTO_DIR \
      $PROTO_DIR/$var
```

- **go_out fastjsonpb** 输出目录要保持一致

### 调用

```
...
import(
  fastjsonpb "github.com/superjsf2010/protoc-gen-fastjsonpb/encoding/json"
  // proto编译产出
  "github.com/superjsf2010/protoc-gen-fastjsonpb/test/example"
)
...
// 序列化
e1 := &example.Example{}
ret,err := fastjsonpb.Marshal(e1)

// 反序列化
e2 := example.ExampleNew()
// 也可以使用e2 := &example.Example{} 创建对象，上述方式会使用Pool提高性能
fastjsonpb.Unmarshal(ret, e2)
// !!!! 为提供性能，需要手动释放对象，释放的对象会进入Pool，当然你也可以不这么做
e2.Destructor()
...
```
