package jsonparser

import (
	"fmt"
	"strconv"

	"github.com/superjsf2010/protoc-gen-fastjsonpb/x/buffer"
)

const (
	EnumUnknown int = iota
	EnumString
	EnumNumber
)

type tokenKind byte

const (
	tokenUnknown tokenKind = iota
	tokenSymbol
	tokenBool
	tokenString
	tokenNumber
	tokenNull
)

type token struct {
	kind   tokenKind
	bol    bool
	raw    []byte
	symbol byte
}

type Parser struct {
	data   []byte
	off    int
	token  *token
	assert byte
}

func New(data []byte) *Parser {
	return &Parser{
		data: data,
		token: &token{
			kind: tokenUnknown,
		},
	}
}

// 获取token
func (p *Parser) getToken() {
	for _, c := range p.data[p.off:] {
		switch c {
		case '{', '[':
			if p.assert != 0 {
				p.syntaxErr()
			}
			p.token.kind = tokenSymbol
			p.token.symbol = c
			p.off++
			return
		case '}', ']':
			if p.assert != ',' {
				p.syntaxErr()
			}
			p.AssertSymbol(0)
			p.token.kind = tokenSymbol
			p.token.symbol = c
			p.off++
			return
		case ' ', '\t', '\r', '\n':
			p.off++
		case '"':
			if p.assert != 0 {
				p.syntaxErr()
			}
			p.off++
			p.token.kind = tokenString
			p.getString()
			return
		case ':', ',':
			if p.assert != c {
				p.syntaxErr()
			}
			p.reset()
			p.AssertSymbol(0)
			p.off++
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			if p.assert != 0 {
				p.syntaxErr()
			}
			p.token.kind = tokenNumber
			p.getNumber()
			return
		case 't':
			if p.assert != 0 {
				p.syntaxErr()
			}
			p.token.kind = tokenBool
			p.getTrue()
			return
		case 'f':
			if p.assert != 0 {
				p.syntaxErr()
			}
			p.token.kind = tokenBool
			p.getFalse()
			return
		case 'n':
			if p.assert != 0 {
				p.syntaxErr()
			}
			p.token.kind = tokenNull
			p.getNull()
			return
		default:
			p.syntaxErr()
		}
	}
}

func (p *Parser) getString() {
	i := p.off
Switch:
	for ; i < len(p.data); i++ {
		switch p.data[i] {
		case '\\':
			// TODO 转码
			i++
			p.escape()
		case '"':
			break Switch
		}
	}
	p.token.raw = p.data[p.off:i]
	// 跳过"
	p.off = i + 1
}

func (p *Parser) escape() {
}

func (p *Parser) getTrue() {
	if p.off+3 > len(p.data) ||
		p.data[p.off+1] != 'r' ||
		p.data[p.off+2] != 'u' ||
		p.data[p.off+3] != 'e' {
		p.syntaxErr()
	}
	p.token.bol = true
	p.off += 4
}

func (p *Parser) getFalse() {
	if p.off+4 > len(p.data) ||
		p.data[p.off+1] != 'a' ||
		p.data[p.off+2] != 'l' ||
		p.data[p.off+3] != 's' ||
		p.data[p.off+4] != 'e' {
		p.syntaxErr()
	}
	p.token.bol = false
	p.off += 5
}

func (p *Parser) getNumber() {
	i := p.off
Switch:
	for ; i < len(p.data); i++ {
		switch p.data[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', 'E', '+', '-':
		default:
			break Switch
		}
	}
	p.token.raw = p.data[p.off:i]
	p.off = i
}

func (p *Parser) getNull() {
	if p.off+3 > len(p.data) ||
		p.data[p.off+1] != 'u' ||
		p.data[p.off+2] != 'l' ||
		p.data[p.off+3] != 'l' {
		p.syntaxErr()
	}
	p.off += 4
}

// 解析string，包括key，value
func (p *Parser) Str() string {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenString {
		p.syntaxErr()
	}
	p.reset()
	return buffer.Bytes2Str(p.token.raw)
}

// 解析boolean
func (p *Parser) Bol() bool {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenBool {
		p.syntaxErr()
	}
	p.reset()
	return p.token.bol
}

// 解析数字，全部装换成float64
func (p *Parser) Number() float64 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseFloat(buffer.Bytes2Str(p.token.raw), 64)
	if err != nil {
		panic(err)
	}
	p.reset()
	return n
}

// 解析int32
func (p *Parser) Int32() int32 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseInt(buffer.Bytes2Str(p.token.raw), 10, 32)
	if err != nil {
		panic(err)
	}
	p.reset()
	return int32(n)
}

// 解析int64
func (p *Parser) Int64() int64 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseInt(buffer.Bytes2Str(p.token.raw), 10, 64)
	if err != nil {
		panic(err)
	}
	p.reset()
	return n
}

// 解析uint32
func (p *Parser) Uint32() uint32 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseUint(buffer.Bytes2Str(p.token.raw), 10, 32)
	if err != nil {
		panic(err)
	}
	p.reset()
	return uint32(n)
}

// 解析uint64
func (p *Parser) Uint64() uint64 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseUint(buffer.Bytes2Str(p.token.raw), 10, 64)
	if err != nil {
		panic(err)
	}
	p.reset()
	return n
}

// 解析float32
func (p *Parser) Float32() float32 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseFloat(buffer.Bytes2Str(p.token.raw), 32)
	if err != nil {
		panic(err)
	}
	p.reset()
	return float32(n)
}

// 解析float64
func (p *Parser) Float64() float64 {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNumber {
		p.syntaxErr()
	}
	n, err := strconv.ParseFloat(buffer.Bytes2Str(p.token.raw), 64)
	if err != nil {
		panic(err)
	}
	p.reset()
	return n
}

// 解析bytes
func (p *Parser) Bytes() []byte {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenString {
		p.syntaxErr()
	}
	p.reset()
	return p.token.raw
}

// 解析enum，兼容数字、字符串两种
func (p *Parser) Enum() (int, string, int32) {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	switch p.token.kind {
	case tokenString:
		return EnumString, p.Str(), 0
	case tokenNumber:
		return EnumNumber, "", p.Int32()
	default:
		p.syntaxErr()
	}
	return EnumUnknown, "", 0
}

// 解析null
func (p *Parser) Null() interface{} {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenNull {
		p.syntaxErr()
	}
	p.reset()
	return nil
}

func (p *Parser) Symbol(b byte) {
	if p.token.kind == tokenUnknown {
		p.getToken()
	}
	if p.token.kind != tokenSymbol || p.token.symbol != b {
		p.syntaxErr()
	}
	p.reset()
}

// 解析对象
func (p *Parser) obj() map[string]interface{} {
	ret := map[string]interface{}{}
	for !p.IsSymbol('}') {
		key := p.Str()
		p.AssertSymbol(':')
		ret[key] = p.Parse()
		p.AssertSymbol(',')
	}
	// 处理空对象情况
	p.AssertSymbol(',')
	p.Symbol('}')
	return ret
}

// 解析数组
func (p *Parser) arr() []interface{} {
	ret := []interface{}{}
	for !p.IsSymbol(']') {
		ret = append(ret, p.Parse())
		p.AssertSymbol(',')
	}
	// 处理空数组情况
	p.AssertSymbol(',')
	p.Symbol(']')
	return ret
}

func (p *Parser) IsSymbol(b byte) bool {
	return p.data[p.off] == b
}

func (p *Parser) reset() {
	p.token.kind = tokenUnknown
}

func (p *Parser) AssertSymbol(b byte) {
	p.assert = b
}

func (p *Parser) syntaxErr() {
	if p.off < len(p.data) {
		panic(`syntax error: near col ` + strconv.FormatInt(int64(p.off), 10) + ` "` + string(p.data[p.off:]) + `"`)
	}
	panic(`syntax error: near col ` + strconv.FormatInt(int64(p.off), 10))
}

func (p *Parser) Parse() interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	// 尝试获取新token，需要清空上一次token信息
	if p.token.kind != tokenUnknown {
		panic("system exception")
	}

	p.getToken()

	switch p.token.kind {
	case tokenBool:
		return p.Bol()
	case tokenString:
		return p.Str()
	case tokenNumber:
		return p.Number()
	case tokenNull:
		return p.Null()
	case tokenSymbol:
		p.reset()
		if p.token.symbol == '{' {
			return p.obj()
		} else if p.token.symbol == '[' {
			return p.arr()
		}
	}

	panic("unknown token")
}
