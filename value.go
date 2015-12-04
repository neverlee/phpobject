package phpobject

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var patNumber, _ = regexp.Compile(`^-?[1-9][0-9]*$`)
var patVarName, _ = regexp.Compile(`^[[:alpha:]_]\w*$`)

type PValueType int

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
	Marshal() []byte
	//Unmarshal(bcode []byte) error
	//unserialize(r io.Reader) error
}

type PNilType struct{}

func (nl *PNilType) String() string   { return "nil" }
func (nl *PNilType) Type() PValueType { return PTNil }
func (nl *PNilType) Marshal() []byte  { return []byte("N;") }
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
func (bl PBool) Marshal() []byte {
	if bl {
		return []byte("b:1;")
	} else {
		return []byte("b:0;")
	}
}
func (bl PBool) serialize(w io.Writer) {
	w.Write(bl.Marshal())
}

var PTrue = PBool(true)
var PFalse = PBool(false)

type PLong int

func (lt PLong) String() string   { return fmt.Sprint(int(lt)) }
func (lt PLong) Type() PValueType { return PTLong }
func (lt PLong) Marshal() []byte {
	return []byte(fmt.Sprintf("i:%d;", lt))
}
func (lt PLong) serialize(w io.Writer) {
	fmt.Fprintf(w, "i:%d;", lt)
}

type PDouble float64

func (dt PDouble) String() string   { return fmt.Sprint(float64(dt)) }
func (dt PDouble) Type() PValueType { return PTDouble }
func (dt PDouble) Marshal() []byte {
	return []byte(fmt.Sprintf("d:%f;", dt))
}
func (dt PDouble) serialize(w io.Writer) {
	fmt.Fprintf(w, "d:%f;", dt)
}

type PString string

func (st PString) String() string   { return string(st) }
func (st PString) Type() PValueType { return PTString }
func (st PString) Marshal() []byte {
	return []byte(fmt.Sprintf("s:%d:\"%s\";", len(st), st))
}
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
func (tb *PArray) Get(key string) (PValue, bool) {
	v, o := tb.array[key]
	return v, o
}

func (tb *PArray) Iset(index int, value PValue) {
	key := fmt.Sprintf("%d", index)
	tb.array[key] = value
}
func (tb *PArray) Set(key string, value PValue) bool {
	tb.array[key] = value
	if key == "0" || patNumber.MatchString(key) {
		return true
	} else {
		return false
	}
}

func serializeKey(w io.Writer, key string) {
	if key == "0" || patNumber.MatchString(key) {
		fmt.Fprintf(w, "i:%s;", key)
	} else {
		fmt.Fprintf(w, "s:%d:\"", len(key))
		fmt.Fprint(w, key)
		fmt.Fprint(w, "\";")
	}
}

func (tb *PArray) String() string {
	slist := make([]string, len(tb.array)+2)
	slist[0] = fmt.Sprintf("Array(%d) [", len(tb.array))
	i := 1
	for k, v := range tb.array {
		slist[i] = fmt.Sprintf("%s : %s,", k, v.String())
		i++
	}
	slist[i] = "]"
	return strings.Join(slist, " ")
}
func (tb *PArray) Type() PValueType { return PTArray }
func (tb *PArray) Marshal() []byte {
	buf := bytes.NewBuffer(nil)
	tb.serialize(buf)
	return buf.Bytes()
}
func (tb *PArray) serialize(w io.Writer) {
	fmt.Fprintf(w, "a:%d:{", len(tb.array))
	for k, v := range tb.array {
		serializeKey(w, k)
		v.serialize(w)
	}
	w.Write([]byte("}"))
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
}

func NewObject(class string) *PObject {
	var ot PObject
	ot.vars = make(map[string]oValue)
	ot.class = class
	return &ot
}

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
	ot.vars[varname] = oValue{value, vtype}
	return nil
}

func (ot *PObject) SetPublicVar(varname string, value PValue) error {
	return ot.SetVar(varname, PublicVar, value)
}

func (ot *PObject) SetProtectedVar(varname string, value PValue) error {
	return ot.SetVar(varname, ProtectedVar, value)
}

func (ot *PObject) SetPrivateVar(varname string, value PValue) error {
	return ot.SetVar(varname, PrivateVar, value)
}

func (ot *PObject) SetBaseVar(clsname, varname string, value PValue) error {
	if !patVarName.MatchString(varname) {
		return errors.New("Error varname")
	}
	if !patVarName.MatchString(clsname) {
		return errors.New("Error class name")
	}
	key := fmt.Sprintf("\x00%s\x00%s", clsname, varname)
	ot.vars[key] = oValue{value, BasePrivateVar}
	return nil
}

func (ot *PObject) GetVar(varname string) (value PValue, vtype int, ok bool) {
	oval, ok := ot.vars[varname]
	return oval.value, oval.varType, ok
}

func (ot *PObject) GetBaseVar(clsname, varname string) (value PValue, ok bool) {
	key := fmt.Sprintf("\x00%s\x00%s", clsname, varname)
	oval, ok := ot.vars[key]
	return oval.value, ok
}

func (ot *PObject) String() string {
	slist := make([]string, len(ot.vars)+2)
	slist[0] = fmt.Sprintf("Object(%s:%d) {", ot.class, len(ot.vars))
	i := 1
	for k, v := range ot.vars {
		switch v.varType {
		case PublicVar:
			slist[i] = fmt.Sprintf("%s : %s,", k, v.value.String())
		case ProtectedVar:
			slist[i] = fmt.Sprintf("-%s : %s,", k, v.value.String())
		case PrivateVar:
			slist[i] = fmt.Sprintf("*%s : %s,", k, v.value.String())
		case BasePrivateVar:
			kk := strings.Replace(k, "\x00", "*", 2)
			slist[i] = fmt.Sprintf("%s : %s,", kk[1:], v.value.String())
		}
		i++
	}
	slist[i] = "}"
	return strings.Join(slist, " ")
}
func (ot *PObject) Type() PValueType { return PTObject }
func (ot *PObject) Marshal() []byte {
	buf := bytes.NewBuffer(nil)
	ot.serialize(buf)
	return buf.Bytes()
}

func (ot *PObject) serialize(w io.Writer) {
	fmt.Fprintf(w, "O:%d:\"%s\"", len(ot.class), ot.class)
	fmt.Fprintf(w, ":%d:{", len(ot.vars))
	for k, v := range ot.vars {
		key := PString(k)
		switch v.varType {
		case ProtectedVar:
			key = PString("\x00*\x00" + k)
		case PrivateVar:
			key = PString(fmt.Sprintf("\x00%s\x00%s", ot.class, k))
			//case PublicVar, BasePrivateVar:
			//	key = k
		}
		key.serialize(w)
		v.value.serialize(w)
	}
	w.Write([]byte("}"))
}
