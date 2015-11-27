package phpobject

import (
	"fmt"
	"io"
)

type PValueType int

//r    zval *   资源（文件指针，数据库连接等）
//z    zval *   无任何操作的zval
const (
	PTNil PValueType = iota
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
	//unserialize(r io.Reader)
}

type PNilType struct{}

func (nl *PNilType) String() string        { return "nil" }
func (nl *PNilType) Type() PValueType      { return PTNil }
func (nl *PNilType) serialize(w io.Writer) {}

var PNil = PValue(&PNilType{})

type PBool bool

func (bl PBool) String() string {
	if bool(bl) {
		return "true"
	}
	return "false"
}
func (bl PBool) Type() PValueType { return PTBool }
func (bl PBool) serialize(w io.Writer) {
	if bl {
		w.Write([]byte("b:1;"))
	} else {
		w.Write([]byte("b:0;"))
	}
}

//if isinstance(obj, basestring):
//    encoded_obj = obj
//    if isinstance(obj, unicode):
//        encoded_obj = obj.encode(charset, errors)
//    s = BytesIO()
//    s.write(b's:')
//    s.write(str(len(encoded_obj)).encode('latin1'))
//    s.write(b':"')
//    s.write(encoded_obj)
//    s.write(b'";')
//    return s.getvalue()
//if isinstance(obj, (list, tuple, dict)):
//    out = []
//    if isinstance(obj, dict):
//        iterable = obj.items()
//    else:
//        iterable = enumerate(obj)
//    for key, value in iterable:
//        out.append(_serialize(key, True))
//        out.append(_serialize(value, False))
//    return b''.join([
//        b'a:',
//        str(len(obj)).encode('latin1'),
//        b':{',
//        b''.join(out),
//        b'}'
//    ])
//if isinstance(obj, phpobject):
//    return b'O' + _serialize(obj.__name__, True)[1:-1] + \
//           _serialize(obj.__php_vars__, False)[1:]
//if object_hook is not None:
//    return _serialize(object_hook(obj), False)
//raise TypeError('can\'t serialize %r' % type(obj))

var PTrue = PBool(true)
var PFalse = PBool(false)

type PLong int

func (lt PLong) String() string   { return fmt.Sprint(lt) }
func (lt PLong) Type() PValueType { return PTLong }
func (lt PLong) serialize(w io.Writer) {
	fmt.Fprintf(w, "i:%d;", lt)
}

type PDouble float64

func (dt PDouble) String() string   { return fmt.Sprint(dt) }
func (dt PDouble) Type() PValueType { return PTDouble }
func (dt PDouble) serialize(w io.Writer) {
	fmt.Fprintf(w, "d:%f;", dt)
}

type PString string

func (st PString) String() string        { return string(st) }
func (st PString) Type() PValueType      { return PTString }
func (dt PDouble) serialize(w io.Writer) {}

const (
	NumArray = 1
	KeyArray = 2
)

type PArray struct {
	array     map[PString]PValue
	forceType int
}

func (tb *PArray) String() string        { return fmt.Sprintf("table: %p", tb) }
func (tb *PArray) Type() PValueType      { return PTArray }
func (dt PDouble) serialize(w io.Writer) {}
