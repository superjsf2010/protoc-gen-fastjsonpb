package buffer

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"
	"unsafe"
)

var BufPool sync.Pool

// TODO 动态扩容
type Buffer struct {
	buf []byte
}

func New() *Buffer {
	if v := BufPool.Get(); v != nil {
		b := v.(*Buffer)
		b.reset()
		return b
	}
	return &Buffer{
		buf: []byte(""),
	}
}

func (b *Buffer) reset() {
	b.buf = b.buf[:0]
}

func (b *Buffer) WriteStr(data string) (int, error) {
	m, err := b.grow(len(data))
	if err == nil {
		return copy(b.buf[m:], data), nil
	}
	return 0, err
}

func (b *Buffer) WriteBool(data bool) (int, error) {
	return b.WriteStr(strconv.FormatBool(data))
}

func (b *Buffer) WriteInt32(data int32) (int, error) {
	return b.WriteStr(strconv.FormatInt(int64(data), 10))
}

func (b *Buffer) WriteInt64(data int64) (int, error) {
	return b.WriteStr(strconv.FormatInt(data, 10))
}

func (b *Buffer) WriteUint32(data uint32) (int, error) {
	return b.WriteStr(strconv.FormatUint(uint64(data), 10))
}

func (b *Buffer) WriteUint64(data uint64) (int, error) {
	return b.WriteStr(strconv.FormatUint(data, 10))
}

func (b *Buffer) WriteFloat32(data float32) (int, error) {
	return b.WriteStr(strconv.FormatFloat(float64(data), 'f', -1, 32))
}

func (b *Buffer) WriteFloat64(data float64) (int, error) {
	return b.WriteStr(strconv.FormatFloat(data, 'f', -1, 64))
}

// 区别于Write，这里将[]byte序列化到json，需要进行base64编码
func (b *Buffer) WriteBytes(data []byte) (int, error) {
	return b.WriteStr(`"` + base64.StdEncoding.EncodeToString(data) + `"`)
}

func (b *Buffer) WriteByte(data byte) (int, error) {
	m, err := b.grow(1)
	if err == nil {
		b.buf[m] = data
		return 1, nil
	}
	return 0, nil
}

func (b *Buffer) Write(data []byte) (int, error) {
	m, err := b.grow(len(data))
	if err == nil {
		return copy(b.buf[m:], data), nil
	}
	return 0, nil
}

func (b *Buffer) FixSymbol() {
	l := len(b.buf)
	if l >= 2 && string(b.buf[l-1:]) == "," {
		b.buf = b.buf[:l-1]
	}
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

func (b *Buffer) grow(n int) (int, error) {
	l := len(b.buf)
	if l+n > cap(b.buf) {
		// TODO grow
		buf := make([]byte, 2*(l+n))
		copy(buf, b.buf)
		b.buf = buf
	}
	b.buf = b.buf[:l+n]
	return l, nil
}

func (b *Buffer) Grow(n int) {
	buf := make([]byte, n)
	copy(buf, b.buf)
	b.buf = buf
}

func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Str2Bytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

const hex = "0123456789abcdef"

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

func (b *Buffer) WriteStringWithQuote(s string) {
	b.WriteByte('"')
	b.WriteString(s)
	b.WriteByte('"')
}

func (b *Buffer) WriteString(s string) {
	start := 0
	for i := 0; i < len(s); {
		if c := s[i]; c < utf8.RuneSelf {
			if safeSet[c] {
				i++
				continue
			}
			if start < i {
				b.WriteStr(s[start:i])
			}
			b.WriteByte('\\')
			switch c {
			case '\\', '"':
				b.WriteByte(c)
			case '\n':
				b.WriteByte('n')
			case '\r':
				b.WriteByte('r')
			case '\t':
				b.WriteByte('t')
			default:
				b.WriteStr(`u00`)
				b.WriteByte(hex[c>>4])
				b.WriteByte(hex[c&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				b.WriteStr(s[start:i])
			}
			b.WriteStr(`\ufffd`)
			i += size
			start = i
			continue
		}
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				b.WriteStr(s[start:i])
			}
			b.WriteStr(`\u202`)
			b.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		b.WriteStr(s[start:])
	}
}
