package buffer

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"sync"
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

func (b *Buffer) WriteString(data string) (int, error) {
	m, err := b.grow(len(data))
	if err == nil {
		return copy(b.buf[m:], data), nil
	}
	return 0, err
}

func (b *Buffer) WriteBool(data bool) (int, error) {
	return b.WriteString(strconv.FormatBool(data))
}

func (b *Buffer) WriteInt32(data int32) (int, error) {
	return b.WriteString(strconv.FormatInt(int64(data), 10))
}

func (b *Buffer) WriteInt64(data int64) (int, error) {
	return b.WriteString(strconv.FormatInt(data, 10))
}

func (b *Buffer) WriteUint32(data uint32) (int, error) {
	return b.WriteString(strconv.FormatUint(uint64(data), 10))
}

func (b *Buffer) WriteUint64(data uint64) (int, error) {
	return b.WriteString(strconv.FormatUint(data, 10))
}

func (b *Buffer) WriteFloat32(data float32) (int, error) {
	return b.WriteString(strconv.FormatFloat(float64(data), 'f', -1, 32))
}

func (b *Buffer) WriteFloat64(data float64) (int, error) {
	return b.WriteString(strconv.FormatFloat(data, 'f', -1, 64))
}

// 区别于Write，这里将[]byte序列化到json，需要进行base64编码
func (b *Buffer) WriteBytes(data []byte) (int, error) {
	return b.WriteString(`"` + base64.StdEncoding.EncodeToString(data) + `"`)
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
