package phpobject

import (
	"fmt"
	"io"
	"regexp"
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

func (nl *PNilType) String() string   { return "nil" }
func (nl *PNilType) Type() PValueType { return PTNil }
func (nl *PNilType) serialize(w io.Writer) {
	w.Write([]byte("N;"))
}

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

func (st PString) String() string   { return string(st) }
func (st PString) Type() PValueType { return PTString }
func (st PString) serialize(w io.Writer) {
	fmt.Fprintf(w, "s:%d:\"", len(st))
	fmt.Fprint(w, st)
	fmt.Fprint(w, "\";")
}

const (
	NumArray = 1
	KeyArray = 2
)

type PArray struct {
	array map[string]PValue
	//forceType int
}

func NewArray() *PArray {
	var at PArray
	at.array = make(map[string]PValue)
	return &at
}

func (tb *PArray) Iget(index int) (PValue, bool) {
	key := fmt.Sprintf("%d", index)
	v, o := tb.array[key]
	return v, o
}
func (tb *PArray) Rget(key string) (PValue, bool) {
	v, o := tb.array[key]
	return v, o
}
func (tb *PArray) Pget(key string) (PValue, bool) {
	v, o := tb.array[key]
	return v, o
}

func (tb *PArray) Iset(index int, value PValue) {
	key := fmt.Sprintf("%d", index)
	tb.array[key] = value
}
func (tb *PArray) Rset(key string, value PValue) {
	tb.array[key] = value
}
func (tb *PArray) Pset(key string, value PValue) string {
	tb.array[key] = value
	return key
}

var patNumber, _ = regexp.Compile(`^-?[1-9][0-9]*$`)

func serializeKey(w io.Writer, key string) {
	if key == "0" || patNumber.MatchString(key) {
		fmt.Fprintf(w, "i:%s;", key)
	} else {
		fmt.Fprintf(w, "s:%d:\"", len(key))
		fmt.Fprint(w, key)
		fmt.Fprint(w, "\";")
	}
}

func (tb *PArray) String() string   { return fmt.Sprintf("table: %v", tb) }
func (tb *PArray) Type() PValueType { return PTArray }
func (tb *PArray) serialize(w io.Writer) {
	fmt.Fprintf(w, "a:%d:{", len(tb.array))
	for k, v := range tb.array {
		serializeKey(w, k)
		v.serialize(w)
	}
	w.Write([]byte("}"))
}
func (tb *PArray) Output(w io.Writer) {
	tb.serialize(w)
}

const (
	PublicVar      = 0
	ProtectedVar   = 1
	PrivateVar     = 2
	BasePrivateVar = 4
	endVarType     = 5
)

type oValue struct {
	value   PValue
	varType int
}

type PObject struct {
	vars  map[string]oValue
	class string
	//forceType int
}

func NewObject(class string) *PObject {
	var ot PObject
	ot.vars = make(map[string]oValue)
	ot.class = class
	return &ot
}

var patVarName, _ = regexp.Compile(`^[[:alpha:]_]\w*$`)

func (ot *PObject) SetVar(varname string, vtype int, value PValue) error {
	if vtype == BasePrivateVar {
		return errors.New("You should use SetBaseVar")
	}
	if vtype > BasePrivateVar || vtype < 0 {
		return errors.New("Error var type")
	}
	if !patVarName.MatchString(varname) {
		return errors.New("Error varname")
	}
	ot.vars[string] = oValue{value, vtype}
}

func (ot *PObject) SetBaseVar(claname, varname string, value PValue) error {
	if !patVarName.MatchString(varname) {
		return errors.New("Error varname")
	}
	if !patVarName.MatchString(clsname) {
		return errors.New("Error class name")
	}
	ot.vars[string] = oValue{value, BasePrivateVar}
}

func (ot *PObject) GetVar(varname string) (value PValue, vtype int, ok bool) {
	oval, ok := ot.vars[varname]
	return oval.value, varType, ok
}

func (ot *PObject) GetBaseVar(clsname, varname string) (value PValue, ok bool) {
	key := fmt.Sprintf("\x00%s\x00%s", clsname, varname)
	oval, ok := ot.vars[varname]
	return oval.value, ok
}

func (ot *PObject) String() string   { return fmt.Sprintf("object: %v", ot) }
func (ot *PObject) Type() PValueType { return PTObject }
func (ot *PObject) serialize(w io.Writer) {
	fmt.Fprintf(w, "O:%d:\"%s\"", len(st), ot.class)
	fmt.Fprintf(w, ":%d:{", len(tb.vars))
	for k, v := range tb.vars {
		key = PString(k)
		switch v.varType {
		case ProtectedVar:
			key = PString("\x00*\x00" + k)
		case PrivateVar:
			key = PString(fmt.Sprintf("\x00%s\x00%s", ot.class, k))
			//case PublicVar, BasePrivateVar:
			//	key = k
		}
		key.serialize(w)
		v.serialize(w)
	}
	w.Write([]byte("}"))
}
