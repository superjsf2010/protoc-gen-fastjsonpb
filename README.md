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

### 性能对比

#### 平台

goos: linux
goarch: amd64
cpu: Intel(R) Xeon(R) CPU E5-26xx v4

#### 源代码

[https://github.com/superjsf2010/protoc-gen-fastjsonpb/blob/main/test/marshal_test.go](https://github.com/superjsf2010/protoc-gen-fastjsonpb/blob/main/test/marshal_test.go)
[https://github.com/superjsf2010/protoc-gen-fastjsonpb/blob/main/test/unmarshal_test.go](https://github.com/superjsf2010/protoc-gen-fastjsonpb/blob/main/test/unmarshal_test.go)

#### 数据

| | ns/op | allocation bytes | allocation times |
| -- | -- | -- | -- |
| FastJsonpb Marshal | 2636 ns/op | 160 B/op	| 12 allocs/op |
| StdJsonpb Marshal | 13927 ns/op | 2024 B/op | 43 allocs/op |
| StdJson Marshal | 3279 ns/op | 384 B/op | 1 allocs/op |
| Jsoniter Marshal | 5541 ns/op | 2528 B/op | 27 allocs/op |
| FastJsonpb Unmarshal | 3523 ns/op | 464 B/op | 7 allocs/op |
| StdJsonpb Unmarshal | 19182 ns/op | 2864 B/op | 101 allocs/op |
| StdJson Unmarshal | 10249 ns/op | 1416 B/op | 18 allocs/op |
| JsoniterUnmarshal | 3354 ns/op | 1672 B/op | 30 allocs/op |

#### 备注

- std-json, jsoniter不支持pb
- 官方标准库已经迭代过多个版本，性能也得到大幅提升
