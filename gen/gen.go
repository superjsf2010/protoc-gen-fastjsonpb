package gen

import (
	"strconv"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

type FastJsonpbGen struct {
	plugin      *protogen.Plugin
	messagesMap map[string]*protogen.Message
	enumsMap    map[string]*protogen.Enum
}

func New(req *pluginpb.CodeGeneratorRequest) (*FastJsonpbGen, error) {
	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		return nil, err
	}
	gen := &FastJsonpbGen{
		plugin:      plugin,
		messagesMap: make(map[string]*protogen.Message),
		enumsMap:    make(map[string]*protogen.Enum),
	}
	return gen, nil
}

func (g *FastJsonpbGen) GenerateAllFiles() (*pluginpb.CodeGeneratorResponse, error) {
	for _, protoFile := range g.plugin.Files {
		filename := protoFile.GeneratedFilenamePrefix + ".pb.fastjsonpb.go"
		gf := g.plugin.NewGeneratedFile(filename, ".")
		g.generateComments(protoFile, gf)
		g.generatePackageName(protoFile, gf)
		g.generateImport(protoFile, gf)
		g.buildMessageIndex(protoFile, gf)
		g.buildEnumIndex(protoFile, gf)
		g.generateMessage(protoFile, gf)
	}
	return g.plugin.Response(), nil
}

// 生成头部注释
func (g *FastJsonpbGen) generateComments(protoFile *protogen.File, gf *protogen.GeneratedFile) {
	gf.P(` // Code generated by protoc-gen-fastjsonpb. DO NOT EDIT.`)
	gf.P(` // source:` + *protoFile.Proto.Name)
	gf.P(``)
}

// 生成包名称
func (g *FastJsonpbGen) generatePackageName(protoFile *protogen.File, gf *protogen.GeneratedFile) {
	gf.P(`package ` + protoFile.GoPackageName)
}

// 生成依赖包
func (g *FastJsonpbGen) generateImport(protoFile *protogen.File, gf *protogen.GeneratedFile) {
	g.genImport("", "github.com/superjsf2010/protoc-gen-fastjsonpb/x/buffer", gf)
	g.genImport("", "github.com/superjsf2010/protoc-gen-fastjsonpb/x/jsonparser", gf)
	g.genImport("", "sync", gf)
}

func (g *FastJsonpbGen) genImport(name string, importPath string, gf *protogen.GeneratedFile) {
	gf.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

func (g *FastJsonpbGen) buildMessageIndex(protoFile *protogen.File, gf *protogen.GeneratedFile) {
	for _, message := range protoFile.Messages {
		// 顶层message
		g.addMessageType(message)
	}
}

func (g *FastJsonpbGen) addMessageType(message *protogen.Message) {
	fullName := string(message.Desc.FullName())
	if _, ok := g.messagesMap[fullName]; !ok {
		g.messagesMap[fullName] = message
	}
	// 嵌套message
	for _, m := range message.Messages {
		g.addMessageType(m)
	}
}

func (g *FastJsonpbGen) buildEnumIndex(protoFile *protogen.File, gf *protogen.GeneratedFile) {
	// 顶层enums
	for _, e := range protoFile.Enums {
		g.addTopEnumType(e)
	}
	// 顶层message
	for _, message := range protoFile.Messages {
		// 嵌套enum
		g.addNestedEnumType(message)
	}
}

func (g *FastJsonpbGen) addTopEnumType(e *protogen.Enum) {
	fullName := string(e.Desc.FullName())
	if _, ok := g.enumsMap[fullName]; !ok {
		g.enumsMap[fullName] = e
	}
}

func (g *FastJsonpbGen) addNestedEnumType(message *protogen.Message) {
	for _, e := range message.Enums {
		g.addTopEnumType(e)
	}
	// 嵌套message
	for _, m := range message.Messages {
		g.addNestedEnumType(m)
	}
}

// 生成message对应信息
func (g *FastJsonpbGen) generateMessage(protoFile *protogen.File, gf *protogen.GeneratedFile) {
	// 处理顶层message
	for _, message := range protoFile.Messages {
		g.genMessage(message, gf)
	}
	// 处理顶层enum
	for _, e := range protoFile.Enums {
		g.generateEnum(e, gf)
	}
	// TODO 处理顶层extension
	// TODO 处理顶层service
}

// 递归生成message对应信息
func (g *FastJsonpbGen) genMessage(message *protogen.Message, gf *protogen.GeneratedFile) {
	//g.generateDebug(message, gf)
	g.generateMarshal(message, gf)
	g.generateUnmarshal(message, gf)
	g.generatePool(message, gf)
	g.generateDestructor(message, gf)
	g.generateEmpty(message, gf)
	// 处理内嵌message
	for _, m := range message.Messages {
		if m.Desc.IsMapEntry() {
			continue
		}
		g.genMessage(m, gf)
	}
	// 处理内嵌enum
	for _, e := range message.Enums {
		g.generateEnum(e, gf)
	}
	// TODO 处理内嵌extension
}

// 生成序列化方法
func (g *FastJsonpbGen) generateMarshal(message *protogen.Message, gf *protogen.GeneratedFile) {
	gf.P(`func (x *` + message.GoIdent.GoName + `) FastMarshal(buf *buffer.Buffer) {`)
	gf.P(`if x == nil {`)
	gf.P(`buf.WriteString("{}")`)
	gf.P(`}`)
	g.symbolMarshal(gf, `{`)
	// 处理simple字段
	for _, f := range message.Fields {
		if f.Desc.ContainingOneof() == nil {
			if f.Desc.IsList() {
				g.listMarshal(gf, f)
			} else if f.Desc.IsMap() {
				g.mapMarshal(gf, f)
			} else {
				g.typeMarshal(gf, f)
			}
			gf.P(``)
		}
	}
	// 处理oneof字段
	for _, of := range message.Oneofs {
		gf.P(`if x.` + of.GoName + ` != nil {`)
		prefix := ``
		for i, osf := range of.Fields {
			// oneof 不支持array map
			if i == 0 {
				prefix = `if _, ok := x.Get` + of.GoName + `().`
			} else {
				prefix = `} else if _, ok := x.Get` + of.GoName + `().`
			}
			g.oneofTypeMarshal(gf, osf, prefix)
			if i+1 == len(of.Fields) {
				gf.P(`}`)
			}
			gf.P(``)
		}
		gf.P(`}`)
		gf.P(``)
	}
	// 处理多余的逗号
	g.fixSymbolMarshal(gf)
	g.symbolMarshal(gf, `}`)
	gf.P(`}`)
	gf.P(``)
}

// 处理array
func (g *FastJsonpbGen) listMarshal(gf *protogen.GeneratedFile, f *protogen.Field) {
	gf.P(`if !x.IsEmpty` + f.GoName + `() {`)
	g.keyMarshal(gf, protoreflect.StringKind, `"`+f.Desc.JSONName()+`"`)
	g.symbolMarshal(gf, `[`)
	gf.P(`for i,_ := range x.` + f.GoName + `{`)
	// 为提高性能使用下标形式访问
	g.valMarshal(gf, f.Desc.Kind(), `x.`+f.GoName+`[i]`)
	g.symbolMarshal(gf, `,`)
	gf.P(`}`)
	g.fixSymbolMarshal(gf)
	g.symbolMarshal(gf, `]`)
	g.symbolMarshal(gf, `,`)
	gf.P(`}`)
}

// 处理map
func (g *FastJsonpbGen) mapMarshal(gf *protogen.GeneratedFile, f *protogen.Field) {
	// protobuf官方文档说明map的key_type
	// where the key_type can be any integral or string type (so, any scalar type except for floating point types and bytes). Note that enum is not a valid key_type
	gf.P(`if !x.IsEmpty` + f.GoName + `() {`)
	g.keyMarshal(gf, protoreflect.StringKind, `"`+f.Desc.JSONName()+`"`)
	key := f.Desc.MapKey()
	val := f.Desc.MapValue()
	g.symbolMarshal(gf, `{`)
	gf.P(`for k,_ := range x.` + f.GoName + `{`)
	// 为了让field和map_key使用同一个keyMarshal方法，这里加个特殊判断
	if key.Kind() == protoreflect.StringKind {
		g.keyMarshal(gf, key.Kind(), `k`)
	} else {
		g.keyMarshal(gf, key.Kind(), `k`)
	}
	// 为提高性能使用下标形式访问
	g.valMarshal(gf, val.Kind(), `x.`+f.GoName+`[k]`)
	g.symbolMarshal(gf, `,`)
	gf.P(`}`)
	g.fixSymbolMarshal(gf)
	g.symbolMarshal(gf, `}`)
	g.symbolMarshal(gf, `,`)
	gf.P(`}`)
}

// 处理一般类型
func (g *FastJsonpbGen) typeMarshal(gf *protogen.GeneratedFile, f *protogen.Field) {
	gf.P(`if !x.IsEmpty` + f.GoName + `() {`)
	g.keyMarshal(gf, protoreflect.StringKind, `"`+f.Desc.JSONName()+`"`)
	g.valMarshal(gf, f.Desc.Kind(), `x.Get`+f.GoName+`()`)
	g.symbolMarshal(gf, `,`)
	gf.P(`}`)
}

// 处理oneof一般类型
func (g *FastJsonpbGen) oneofTypeMarshal(gf *protogen.GeneratedFile, f *protogen.Field, prefix string) {
	gf.P(prefix + `(*` + f.GoIdent.GoName + `); ok {`)
	g.keyMarshal(gf, protoreflect.StringKind, `"`+f.Desc.JSONName()+`"`)
	g.valMarshal(gf, f.Desc.Kind(), `x.Get`+f.GoName+`()`)
	g.symbolMarshal(gf, `,`)
}

func (g *FastJsonpbGen) valMarshal(gf *protogen.GeneratedFile, k protoreflect.Kind, v string) {
	switch k {
	case protoreflect.BoolKind:
		gf.P(`buf.WriteBool(` + v + `)`)
	case protoreflect.EnumKind:
		gf.P(`buf.WriteStringWithQuote(` + v + `.String())`)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		gf.P(`buf.WriteInt32(` + v + `)`)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		gf.P(`buf.WriteUint32(` + v + `)`)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		gf.P(`buf.WriteInt64(` + v + `)`)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		gf.P(`buf.WriteUint64(` + v + `)`)
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		gf.P(`buf.WriteFloat32(` + v + `)`)
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		gf.P(`buf.WriteFloat64(` + v + `)`)
	case protoreflect.StringKind:
		gf.P(`buf.WriteStringWithQuote(` + v + `)`)
	case protoreflect.BytesKind:
		gf.P(`buf.WriteBytes(` + v + `)`)
	case protoreflect.MessageKind:
		gf.P(v + `.FastMarshal(buf)`)
	case protoreflect.GroupKind:
		// TODO  unspported type
	default:
		// TODO  unknown type
	}
}

// 写入key
func (g *FastJsonpbGen) keyMarshal(gf *protogen.GeneratedFile, k protoreflect.Kind, key string) {
	switch k {
	case protoreflect.StringKind:
		gf.P(`buf.WriteStringWithQuote(` + key + `)`)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		gf.P(`buf.WriteInt32(` + key + `)`)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		gf.P(`buf.WriteUint32(` + key + `)`)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		gf.P(`buf.WriteInt64(` + key + `)`)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		gf.P(`buf.WriteUint64(` + key + `)`)
	default:
		// TODO unspported type
	}
	g.symbolMarshal(gf, `:`)
}

// 写入界定符号: { } [ ] ,
func (g *FastJsonpbGen) symbolMarshal(gf *protogen.GeneratedFile, symbol string) {
	gf.P(`buf.WriteString("` + symbol + `")`)
}

// 多余逗号处理,
func (g *FastJsonpbGen) fixSymbolMarshal(gf *protogen.GeneratedFile) {
	gf.P(`buf.FixSymbol()`)
}

func (g *FastJsonpbGen) mapKeyTypeName(f *protogen.Field) string {
	key := f.Desc.MapKey()
	switch key.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.StringKind:
		return "string"
	default:
		panic("unknown type")
	}
}

func (g *FastJsonpbGen) mapValTypeName(f *protogen.Field, needStar bool) string {
	val := f.Desc.MapValue()
	switch val.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		if e, ok := g.enumsMap[string(val.Enum().FullName())]; ok {
			return e.GoIdent.GoName
		}
		panic("unknown type:" + string(val.Enum().FullName()))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return "float32"
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		if message, ok := g.messagesMap[string(val.Message().FullName())]; ok {
			if needStar {
				return `*` + message.GoIdent.GoName
			} else {
				return message.GoIdent.GoName
			}
		}
		panic("unknown type:" + string(val.Message().FullName()))
	case protoreflect.GroupKind:
		// TODO  unspported type
	default:
		panic("unknown type")
	}
	panic("unknown type")
}

func (g *FastJsonpbGen) typeName(f *protogen.Field, needStar bool) string {
	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		return f.Enum.GoIdent.GoName
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return "float32"
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		if needStar {
			return `*` + f.Message.GoIdent.GoName
		} else {
			return f.Message.GoIdent.GoName
		}
	default:
		panic("unknown type")
	}
	panic("unknown type")
}

// TODO 删除debug信息
func (g *FastJsonpbGen) generateDebug(message *protogen.Message, gf *protogen.GeneratedFile) {
	if message.Desc.IsMapEntry() {
		return
	}
	for _, sf := range message.Fields {
		if oneof := sf.Desc.ContainingOneof(); oneof == nil {
			str := `// goName:` + sf.GoName
			str += ` name:` + string(sf.Desc.Name())
			str += ` fullName:` + string(sf.Desc.FullName())
			str += ` kind:` + sf.Desc.Kind().String()
			str += ` jsonName:` + sf.Desc.JSONName()
			if sf.Desc.IsMap() {
				str += ` IsMap:true`
				str += ` typeName:` + g.mapValTypeName(sf, false)
				key := sf.Desc.MapKey()
				val := sf.Desc.MapValue()
				str += ` mapKeyKind:` + key.Kind().String()
				if key.Kind().String() == "message" {
					str += ` mapKeyName:` + string(key.Message().Name())
					str += ` mapKeyFullName:` + string(key.Message().FullName())
				}
				str += ` mapValueKind:` + sf.Desc.MapValue().Kind().String()
				if val.Kind().String() == "message" {
					str += ` mapValueName:` + string(val.Message().Name())
					str += ` mapValueFullName:` + string(val.Message().FullName())
				}
			}
			if sf.Desc.IsList() {
				str += ` IsList:true`
				if sf.Desc.Kind().String() == "message" {
					str += ` listValueName:` + string(sf.Desc.Message().Name())
				}
			}
			str += ` IsExtension:` + strconv.FormatBool(sf.Desc.IsExtension())
			gf.P(str)
		}
	}
	for _, of := range message.Oneofs {
		str := `// goName:` + of.GoName
		gf.P(str)
		for _, osf := range of.Fields {
			str := `// goName:` + osf.GoName
			str += ` kind:` + osf.Desc.Kind().String()
			str += ` goident:` + osf.GoIdent.GoName
			///str += ` jsonName:` + osf.Desc.JSONName()
			//str += ` IsExtension:` + strconv.FormatBool(osf.Desc.IsExtension())
			gf.P(str)
		}
	}
	gf.P(``)
}

// 生成反序列化方法
func (g *FastJsonpbGen) generateUnmarshal(message *protogen.Message, gf *protogen.GeneratedFile) {
	gf.P(`func (x *` + message.GoIdent.GoName + `) FastUnmarshal(p *jsonparser.Parser) {`)
	gf.P(`if x == nil {`)
	gf.P(`panic("type ` + message.GoIdent.GoName + ` is nil")`)
	gf.P(`}`)
	gf.P(`p.Symbol('{')`)
	gf.P(`for !p.IsSymbol('}') {`)
	gf.P(`key := p.Str()`)
	gf.P(`p.AssertSymbol(':')`)
	gf.P(`switch key {`)
	// 处理simple字段
	for _, f := range message.Fields {
		if f.Desc.ContainingOneof() == nil {
			if f.Desc.IsList() {
				g.listUnmarshal(gf, f)
			} else if f.Desc.IsMap() {
				g.mapUnmarshal(gf, f)
			} else {
				g.typeUnmarshal(gf, f)
			}
			gf.P(``)
		}
	}
	// 处理oneof字段
	for _, of := range message.Oneofs {
		for _, f := range of.Fields {
			g.oneofTypeUnmarshal(gf, of, f)
		}
	}
	gf.P(`default:`)
	gf.P(`p.PassParse()`)
	// end switch
	gf.P(`}`)
	gf.P(`p.AssertSymbol(',')`)
	// end for
	gf.P(`}`)
	gf.P(`p.AssertSymbol(',')`)
	gf.P(`p.Symbol('}')`)
	// end func
	gf.P(`}`)
	gf.P(``)
}

func (g *FastJsonpbGen) valUnmarshal(gf *protogen.GeneratedFile, k protoreflect.Kind, v string, typeName string) {
	switch k {
	case protoreflect.BoolKind:
		gf.P(v + ` = p.Bol()`)
	case protoreflect.EnumKind:
		gf.P(`t,s,i := p.Enum()`)
		gf.P(`if t == jsonparser.EnumNumber {`)
		gf.P(v + `.Set(i)`)
		gf.P(`}else {`)
		gf.P(v + `.SetByStr(s)`)
		gf.P(`}`)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		gf.P(v + ` = p.Int32()`)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		gf.P(v + ` = p.Uint32()`)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		gf.P(v + ` = p.Int64()`)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		gf.P(v + ` = p.Uint64()`)
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		gf.P(v + ` = p.Float32()`)
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		gf.P(v + ` = p.Float64()`)
	case protoreflect.StringKind:
		gf.P(v + ` = p.Str()`)
	case protoreflect.BytesKind:
		gf.P(v + ` = p.Bytes()`)
	case protoreflect.MessageKind:
		gf.P(v + ` = ` + typeName + `New()`)
		gf.P(v + `.FastUnmarshal(p)`)
	case protoreflect.GroupKind:
		// TODO  unspported type
	default:
		// TODO  unknown type
	}
}

func (g *FastJsonpbGen) typeUnmarshal(gf *protogen.GeneratedFile, f *protogen.Field) {
	gf.P(`case "` + f.Desc.JSONName() + `":`)
	g.valUnmarshal(gf, f.Desc.Kind(), `x.`+f.GoName, g.typeName(f, false))
}

func (g *FastJsonpbGen) listValUnmarshal(gf *protogen.GeneratedFile, k protoreflect.Kind, v string, typeName string) {
	switch k {
	case protoreflect.BoolKind:
		gf.P(v + ` = append(` + v + `,p.Bol())`)
	case protoreflect.EnumKind:
		gf.P(`var e ` + typeName)
		gf.P(`t,s,i := p.Enum()`)
		gf.P(`if t == jsonparser.EnumNumber {`)
		gf.P(`e = e.Get(i)`)
		gf.P(`}else {`)
		gf.P(`e = e.GetByStr(s)`)
		gf.P(`}`)
		gf.P(v + ` = append(` + v + `,e)`)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		gf.P(v + ` = append(` + v + `,p.Int32())`)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		gf.P(v + ` = append(` + v + `,p.Uint32())`)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		gf.P(v + ` = append(` + v + `,p.Int64())`)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		gf.P(v + ` = append(` + v + `,p.Uint64())`)
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		gf.P(v + ` = append(` + v + `,p.Float32())`)
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		gf.P(v + ` = append(` + v + `,p.Float64())`)
	case protoreflect.StringKind:
		gf.P(v + ` = append(` + v + `,p.Str())`)
	case protoreflect.BytesKind:
		gf.P(v + ` = append(` + v + `,p.Bytes())`)
	case protoreflect.MessageKind:
		gf.P(`tmp := ` + typeName + `New()`)
		gf.P(`tmp.FastUnmarshal(p)`)
		gf.P(v + ` = append(` + v + `,tmp)`)
	case protoreflect.GroupKind:
		// TODO  unspported type
	default:
		// TODO  unknown type
	}
}

func (g *FastJsonpbGen) listUnmarshal(gf *protogen.GeneratedFile, f *protogen.Field) {
	gf.P(`case "` + f.Desc.JSONName() + `":`)
	gf.P(`p.Symbol('[')`)
	gf.P(`arr := make([]` + g.typeName(f, true) + `,0)`)
	gf.P(`for !p.IsSymbol(']') {`)
	g.listValUnmarshal(gf, f.Desc.Kind(), `arr`, g.typeName(f, false))
	gf.P(`p.AssertSymbol(',')`)
	//end for
	gf.P(`}`)
	gf.P(`p.AssertSymbol(',')`)
	gf.P(`p.Symbol(']')`)
	gf.P(`x.` + f.GoName + ` = arr`)
}

func (g *FastJsonpbGen) mapValUnmarshal(gf *protogen.GeneratedFile, k protoreflect.Kind, v string, typeName string) {
	switch k {
	case protoreflect.BoolKind:
		gf.P(v + ` = p.Bol()`)
	case protoreflect.EnumKind:
		gf.P(`var e ` + typeName)
		gf.P(`t,s,i := p.Enum()`)
		gf.P(`if t == jsonparser.EnumNumber {`)
		gf.P(`e = e.Get(i)`)
		gf.P(`}else {`)
		gf.P(`e = e.GetByStr(s)`)
		gf.P(`}`)
		gf.P(v + ` = e`)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		gf.P(v + ` = p.Int32()`)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		gf.P(v + ` = p.Uint32()`)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		gf.P(v + ` = p.Int64()`)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		gf.P(v + ` = p.Uint64()`)
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		gf.P(v + ` = p.Float32()`)
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		gf.P(v + ` = p.Float64()`)
	case protoreflect.StringKind:
		gf.P(v + ` = p.Str()`)
	case protoreflect.BytesKind:
		gf.P(v + ` = p.Bytes()`)
	case protoreflect.MessageKind:
		gf.P(`tmp := ` + typeName + `New()`)
		gf.P(`tmp.FastUnmarshal(p)`)
		gf.P(v + ` = tmp`)
	case protoreflect.GroupKind:
		// TODO  unspported type
	default:
		// TODO  unknown type
	}
}

func (g *FastJsonpbGen) mapUnmarshal(gf *protogen.GeneratedFile, f *protogen.Field) {
	gf.P(`case "` + f.Desc.JSONName() + `":`)
	gf.P(`p.Symbol('{')`)
	gf.P(`m := make(map[` + g.mapKeyTypeName(f) + `]` + g.mapValTypeName(f, true) + `)`)
	gf.P(`for !p.IsSymbol('}') {`)
	gf.P(`key := p.Str()`)
	gf.P(`p.AssertSymbol(':')`)
	g.mapValUnmarshal(gf, f.Desc.MapValue().Kind(), `m[key]`, g.mapValTypeName(f, false))
	gf.P(`p.AssertSymbol(',')`)
	//end for
	gf.P(`}`)
	gf.P(`p.AssertSymbol(',')`)
	gf.P(`p.Symbol('}')`)
	gf.P(`x.` + f.GoName + ` = m`)
}

func (g *FastJsonpbGen) oneofTypeUnmarshal(gf *protogen.GeneratedFile, of *protogen.Oneof, f *protogen.Field) {
	gf.P(`case "` + f.Desc.JSONName() + `":`)
	gf.P(`tmp := &` + f.GoIdent.GoName + `{}`)
	g.valUnmarshal(gf, f.Desc.Kind(), `tmp.`+f.GoName, g.typeName(f, false))
	gf.P(`x.` + of.GoName + ` = tmp`)
}

// 生成判空方法
func (g *FastJsonpbGen) generateEmpty(message *protogen.Message, gf *protogen.GeneratedFile) {
	// 处理simple字段
	for _, f := range message.Fields {
		if f.Desc.ContainingOneof() == nil {
			gf.P(`func (x *` + message.GoIdent.GoName + `) IsEmpty` + f.GoName + `() bool {`)
			if f.Desc.IsList() || f.Desc.IsMap() {
				gf.P(`return x.Get` + f.GoName + `() == nil`)
			} else {
				switch f.Desc.Kind() {
				case protoreflect.BoolKind:
					gf.P(`return x.Get` + f.GoName + `() == false`)
				case protoreflect.EnumKind:
					gf.P(`return int32(x.Get` + f.GoName + `()) == 0`)
				case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
					gf.P(`return x.Get` + f.GoName + `() == 0`)
				case protoreflect.StringKind:
					gf.P(`return x.Get` + f.GoName + `() == ""`)
				case protoreflect.BytesKind, protoreflect.MessageKind:
					gf.P(`return x.Get` + f.GoName + `() == nil`)
				case protoreflect.GroupKind:
					// TODO  unspported type
				default:
					// TODO  unknown type
				}
			}
			gf.P(`}`)
			gf.P(``)
		}
	}
	gf.P(``)
}

// 生成Pool方法
func (g *FastJsonpbGen) generatePool(message *protogen.Message, gf *protogen.GeneratedFile) {
	gf.P(`var ` + message.GoIdent.GoName + `Pool sync.Pool`)
	gf.P(`func ` + message.GoIdent.GoName + `New() *` + message.GoIdent.GoName + `{`)
	gf.P(`if v := ` + message.GoIdent.GoName + `Pool.Get();v != nil {`)
	gf.P(`return v.(*` + message.GoIdent.GoName + `)`)
	gf.P(`}`)
	gf.P(`return &` + message.GoIdent.GoName + `{}`)
	gf.P(`}`)
}

// 生成enum相关方法
func (g *FastJsonpbGen) generateEnum(e *protogen.Enum, gf *protogen.GeneratedFile) {
	g.generateEnumGetter(e, gf)
	g.generateEnumSetter(e, gf)
}

func (g *FastJsonpbGen) generateEnumSetter(e *protogen.Enum, gf *protogen.GeneratedFile) {
	goName := e.GoIdent.GoName
	gf.P(`func (x ` + goName + `) Set(i int32) {`)
	gf.P(`if _,ok := ` + goName + `_name[i]; ok {`)
	gf.P(`x = ` + goName + `(i)`)
	gf.P(`return`)
	gf.P(`}`)
	gf.P(`panic("enum ` + goName + `value do not match")`)
	gf.P(`}`)
	gf.P(``)

	gf.P(`func (x ` + goName + `) SetByStr(s string) {`)
	gf.P(`if i,ok := ` + goName + `_value[s]; ok {`)
	gf.P(`x = ` + goName + `(i)`)
	gf.P(`return`)
	gf.P(`}`)
	gf.P(`panic("enum ` + goName + `value do not match")`)
	gf.P(`}`)
	gf.P(``)
}

func (g *FastJsonpbGen) generateEnumGetter(e *protogen.Enum, gf *protogen.GeneratedFile) {
	goName := e.GoIdent.GoName
	gf.P(`func (x ` + goName + `) Get(i int32) ` + goName + `{`)
	gf.P(`if _,ok := ` + goName + `_name[i]; ok {`)
	gf.P(`return ` + goName + `(i)`)
	gf.P(`}`)
	gf.P(`panic("enum ` + goName + `value do not match")`)
	gf.P(`}`)
	gf.P(``)

	gf.P(`func (x ` + goName + `) GetByStr(s string) ` + goName + `{`)
	gf.P(`if i,ok := ` + goName + `_value[s]; ok {`)
	gf.P(`return ` + goName + `(i)`)
	gf.P(`}`)
	gf.P(`panic("enum ` + goName + `value do not match")`)
	gf.P(`}`)
	gf.P(``)
}

// 生成Reset方法
func (g *FastJsonpbGen) generateDestructor(message *protogen.Message, gf *protogen.GeneratedFile) {
	gf.P(`func (x *` + message.GoIdent.GoName + `) Destructor() {`)
	gf.P(`if x == nil {`)
	gf.P(`panic("type ` + message.GoIdent.GoName + ` is nil")`)
	gf.P(`}`)
	// 处理simple字段
	for _, f := range message.Fields {
		if f.Desc.ContainingOneof() == nil {
			if f.Desc.IsList() {
				g.listDestructor(gf, f)
			} else if f.Desc.IsMap() {
				g.mapDestructor(gf, f)
			} else {
				g.typeDestructor(gf, f)
			}
		}
	}
	// 处理oneof字段
	for _, of := range message.Oneofs {
		oneofs := make([]*protogen.Field, 0)
		for _, osf := range of.Fields {
			if osf.Desc.Kind() == protoreflect.MessageKind {
				oneofs = append(oneofs, osf)
			}
		}
		gf.P(`if x.` + of.GoName + ` != nil {`)
		prefix := ``
		for i, osf := range oneofs {
			// oneof 不支持array map
			if i == 0 {
				prefix = `if _, ok := x.Get` + of.GoName + `().`
			} else {
				prefix = `} else if _, ok := x.Get` + of.GoName + `().`
			}
			g.oneofTypeDestructor(gf, of, osf, prefix)
			if i+1 == len(oneofs) {
				gf.P(`}`)
			}
		}
		gf.P(`x.` + of.GoName + ` = nil`)
		gf.P(`}`)
	}
	gf.P(message.GoIdent.GoName + `Pool.Put(x)`)
	// end func
	gf.P(`}`)
	gf.P(``)
}

func (g *FastJsonpbGen) typeDestructor(gf *protogen.GeneratedFile, f *protogen.Field) {
	v := `x.` + f.GoName
	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		gf.P(v + ` = false`)
	case protoreflect.EnumKind:
		gf.P(v + `.Set(0)`)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		gf.P(v + ` = 0`)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		gf.P(v + ` = 0`)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		gf.P(v + ` = 0`)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		gf.P(v + ` = 0`)
	case protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		gf.P(v + ` = 0`)
	case protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		gf.P(v + ` = 0`)
	case protoreflect.StringKind:
		gf.P(v + ` = ""`)
	case protoreflect.BytesKind:
		gf.P(v + ` = nil`)
	case protoreflect.MessageKind:
		gf.P(v + `.Destructor()`)
		gf.P(v + ` = nil`)
	case protoreflect.GroupKind:
		// TODO  unspported type
	default:
		// TODO  unknown type
	}
}

func (g *FastJsonpbGen) listDestructor(gf *protogen.GeneratedFile, f *protogen.Field) {
	if f.Desc.Kind() == protoreflect.MessageKind {
		gf.P(`for i,_ := range x.` + f.GoName + `{`)
		gf.P(`x.` + f.GoName + `[i].Destructor()`)
		gf.P(`}`)
	}
	gf.P(`x.` + f.GoName + ` = nil`)
}

func (g *FastJsonpbGen) mapDestructor(gf *protogen.GeneratedFile, f *protogen.Field) {
	if f.Desc.MapValue().Kind() == protoreflect.MessageKind {
		gf.P(`for i,_ := range x.` + f.GoName + `{`)
		gf.P(`x.` + f.GoName + `[i].Destructor()`)
		gf.P(`}`)
	}
	gf.P(`x.` + f.GoName + ` = nil`)
}

func (g *FastJsonpbGen) oneofTypeDestructor(gf *protogen.GeneratedFile, of *protogen.Oneof, f *protogen.Field, prefix string) {
	gf.P(prefix + `(*` + f.GoIdent.GoName + `); ok {`)
	gf.P(`x.Get` + f.GoName + `().Destructor()`)
}
