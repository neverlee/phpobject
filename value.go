package phpobject

import (
	"fmt"
	"io"
	"os"
)

type PValueType int

//r    zval *   资源（文件指针，数据库连接等）
//z    zval *   无任何操作的zval
const (
	PTNil LValueType = iota
	PTBool
	PTLong
	PTDouble
	PTString
	PTArray
	PTObject
)

var pValueNames = [9]string{"nil", "boolean", "long", "double", "string", "array", "object"}

func (vt PValueType) String() string {
	return pValueNames[int(vt)]
}

type PValue interface {
	String() string
	Type() PValueType
	serialize(w io.Writer)
	unserialize(r io.Reader)
}

type PNilType struct{}

func (nl *PNilType) String() string   { return "nil" }
func (nl *PNilType) Type() PValueType { return PTNil }

var PNil = PValue(&PNilType{})

type PBool bool

func (bl PBool) String() string {
	if bool(bl) {
		return "true"
	}
	return "false"
}
func (bl PBool) Type() PValueType { return PTBool }

var PTrue = PBool(true)
var PFalse = PBool(false)

type PLong int

func (lt PLong) String() string {
	return fmt.Sprint(lt)
}
func (lt PLong) Type() PValueType { return PTLong }

// fmt.Formatter interface
func (lt PLong) Format(f fmt.State, c rune) {
	switch c {
	case 'q', 's':
		defaultFormat(lt.String(), f, c)
	case 'b', 'c', 'd', 'o', 'x', 'X', 'U':
		defaultFormat(int64(lt), f, c)
	case 'e', 'E', 'f', 'F', 'g', 'G':
		defaultFormat(float64(lt), f, c)
	case 'i':
		defaultFormat(int64(lt), f, 'd')
	default:
		defaultFormat(int64(lt), f, c)
	}
}

type PDouble float64

func (dt PDouble) String() string   { return fmt.Sprint(dt) }
func (dt PDouble) Type() PValueType { return PTDouble }

type PString string

func (st PString) String() string   { return string(st) }
func (st PString) Type() PValueType { return PTString }

// fmt.Formatter interface
func (st PString) Format(f fmt.State, c rune) {
	switch c {
	case 'd', 'i':
		if nm, err := parseNumber(string(st)); err != nil {
			defaultFormat(nm, f, 'd')
		} else {
			defaultFormat(string(st), f, 's')
		}
	default:
		defaultFormat(string(st), f, c)
	}
}

const (
	NumArray = 1
	KeyArray = 2
)

type PArray struct {
	array     map[PString]PValue
	forceType int
}

func (tb *PTable) String() string   { return fmt.Sprintf("table: %p", tb) }
func (tb *PTable) Type() PValueType { return PTArray }
