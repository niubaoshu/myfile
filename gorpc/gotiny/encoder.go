package gotiny

import (
	"gorpc/utils"
	"math"
	"reflect"
)

const (
	tooBig = 1 << 30
)

var (
	byte10 = [10]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type encoder struct {
	buff    []byte
	boolBit byte
	boolPos int
}

func NewEncoder() *encoder {
	return &encoder{
		utils.GetBytes()[:0], 0, 0,
	}
}

func (e *encoder) Bytes() []byte {
	return e.buff
}

func Encode(in ...interface{}) []byte {
	return EncodeWithPrefix([]byte{}, in...)
}

func EncodeWithPrefix(space []byte, in ...interface{}) []byte {
	vs := make([]reflect.Value, len(in))
	for i := 0; i < len(in); i++ {
		if reflect.TypeOf(in[i]).Kind() != reflect.Ptr {
			panic("必须是一个指针")
		}
		vs[i] = reflect.ValueOf(in[i]).Elem()
	}
	return EncodeValuesWithPrefix(space, vs...)
}

func EncodeValues(vs ...reflect.Value) []byte {
	return EncodeValuesWithPrefix([]byte{}, vs...)
}

func EncodeValuesWithPrefix(space []byte, vs ...reflect.Value) []byte {
	e := NewEncoder()
	e.buff = append(e.buff, space...)
	for i := 0; i < len(vs); i++ {
		if vs[i].Kind() == reflect.Ptr && vs[i].IsNil() {
			panic("gob: cannot encode nil pointer of type " + vs[i].Type().String())
		}
		e.EncodeValue(vs[i])
	}
	return e.Bytes()
}

func (e *encoder) EncodeValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Bool:
		e.EncBool(v.Bool())
	case reflect.Uint8:
		e.EncUint8(uint8(v.Uint()))
	case reflect.Int8:
		e.EncInt8((int8(v.Int())))
	case reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		e.EncUint(v.Uint())
	case reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		e.EncInt(v.Int())
	case reflect.Float32, reflect.Float64:
		e.EncFloat(v.Float())
	case reflect.Complex64, reflect.Complex128:
		e.EncComplex(v.Complex())
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			e.EncodeValue(v.Index(i))
		}
	case reflect.Map:
		l := v.Len()
		e.EncUint(uint64(l))
		keys := v.MapKeys()
		for i := 0; i < l; i++ {
			e.EncodeValue(keys[i])
			e.EncodeValue(v.MapIndex(keys[i]))
		}
	case reflect.Ptr:
		//if !v.IsNil() {
		e.EncodeValue(v.Elem())
		//}
	case reflect.Slice:
		l := v.Len()
		e.EncUint(uint64(l))
		e.EncUint(uint64(v.Cap()))
		for i := 0; i < l; i++ {
			e.EncodeValue(v.Index(i))
		}
	case reflect.String:
		l := v.Len()
		e.EncUint(uint64(l))
		e.buff = append(e.buff, []byte(v.String())...)
	case reflect.Struct:
		vt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if vt.Field(i).PkgPath == "" { // vt.Field(i).PkgPath 等于 ""代表导出字段
				e.EncodeValue(v.Field(i))
			}
		}
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Invalid:
		//panic("暂不支持这些类型")
	}
}

func (e *encoder) EncBool(v bool) {
	if e.boolBit == 0 {
		e.boolBit = 1
		e.boolPos = len(e.buff)
		e.buff = append(e.buff, 0)
	}
	if v {
		e.buff[e.boolPos] |= e.boolBit
	}
	e.boolBit <<= 1
}

func (e *encoder) EncUint8(v uint8) {
	e.buff = append(e.buff, v)
}

func (e *encoder) EncInt8(v int8) {
	e.buff = append(e.buff, uint8(v))
}

func (e *encoder) EncUint(v uint64) {
	e.buff = append(e.buff, byte10[:putUvarint(byte10[:10], v)]...)
}

func (e *encoder) EncInt(v int64) {
	e.buff = append(e.buff, byte10[:putVarint(byte10[:10], v)]...)
}

func (e *encoder) EncFloat(v float64) {
	e.EncUint(floatBits(v))
}

func (e *encoder) EncComplex(v complex128) {
	e.EncFloat(real(v))
	e.EncFloat(imag(v))
}

func floatBits(f float64) uint64 {
	u := math.Float64bits(f)
	var v uint64
	for i := 0; i < 8; i++ {
		v <<= 8
		v |= u & 0xFF
		u >>= 8
	}
	return v

}
func putUvarint(buf []byte, x uint64) int {
	i := 0
	for x >= 0x80 {
		buf[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	buf[i] = byte(x)
	return i + 1
}

func putVarint(buf []byte, x int64) int {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return putUvarint(buf, ux)
}

// valid reports whether the value is valid and a non-nil pointer.
// (Slices, maps, and chans take care of themselves.)
func valid(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return false
	case reflect.Ptr:
		return !v.IsNil()
	}
	return true
}
