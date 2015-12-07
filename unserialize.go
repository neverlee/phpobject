package phpobject

import (
	"fmt"
	"io"
)

func unserializaNil(r io.Reader) (ret PNilType, err error) {
	var s string
	fmt.Fscanf(r, "%1s", &s)
	if s == ";" {
		return PNil, nil
	}
	return PNil, errors.New("Unserialize Nil fail")
}

func unserializeBool(r io.Reader) (ret PBool, err error) {
	var i int
	if n, nerr := fmt.Fscanf(r, ":%1d;", &i); nerr == nil && n == 1 {
		return PBool(i != 0), nil
	}
	return PFalse, errors.New("Unserialize Bool fail")
}

func unserializeLong(r io.Reader) (ret PLong, err error) {
	var i int
	if n, nerr := fmt.Fscanf(r, ":%d;", &i); nerr == nil && n == 1 {
		return PLong(i), nil
	}
	return 0, errors.New("Unserialize Long fail")
}

func unserializeDouble(r io.Reader) (ret PDouble, err error) {
	var d float64
	if n, nerr := fmt.Fscanf(r, ":%f;", &d); nerr == nil && n == 1 {
		return PDouble(d), nil
	}
	return 0, errors.New("Unserialize Double fail")
}

func unserializeString(r io.Reader) (ret PString, err error) {
	var l int
	if ln, lerr := fmt.Fscanf(r, ":%d:\"", &l); lerr == nil && ln == 1 {
		buf := make([]byte, l+1)
		if bn, berr := io.ReadFull(r, buf); berr == nil && buf[l] == '"' {
			return PString(buf[:l]), nil
		}
	}
	return "", errors.New("Unserialize String fail")
}

func unserializeKey(r io.Reader, isstr bool) (ret string, err error) {
	if isstr {
		var l int
		if ln, lerr := fmt.Fscanf(r, ":%d:\"", &l); lerr == nil && ln == 1 {
			buf := make([]byte, l+1)
			if bn, berr := io.ReadFull(r, buf); berr == nil && buf[l] == '"' {
				return PString(buf[:l]), nil
			}
		}
	} else {
		var s string
		if ln, lerr := fmt.Fscanf(r, ":%[^;];", &l); lerr == nil && ln == 1 {
			return s, nil
		}
	}
	return "", errors.New("Unserialize Key fail")
}

func unserializeArray(r io.Reader) (ret *PArray, err error) {
	var l int
	if ln, lerr := fmt.Fscanf(r, ":%d:{", &l); lerr == nil && ln == 1 {
		array := NewArray()
		for i := 0; i < l; i++ {
			var s string
			fmt.Fscanf(r, "%1s", &s)
			key, kerr := unserializeKey(r, s == 's')
			if kerr != nil {
				return nil, kerr
			}
			val, verr := Unserialize(r)
			if verr != nil {
				return nil, verr
			}
			array.Set(key, val)
		}
		var s string
		fmt.Fscanf(r, "%1s", &s)
		if s == "}" {
			return array, nil
		}
	}
	return "", errors.New("Unserialize Array fail")
}

func unserializeObject(r io.Reader) (ret *PObject, err error) {
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
		key.Serialize(w)
		v.value.Serialize(w)
	}
	w.Write([]byte("}"))

	var cl int
	if ln, lerr := fmt.Fprintf(r, ":%d:\"", &cl); lerr == nil && ln == 1 {
		cnbuf := make([]byte, cl+1)
		var cname string
		if cn, cerr := io.ReadFull(r, cnbuf); cerr == nil && cnbuf[l] == '"' {
			cname = string(cnamebuf[:cl])
		} else {
			return "", errors.New("Unserialize Object Name fail")
		}
		var n int
		if nn, nerr := fmt.Fprintf(r, ":%d:{", &n); nerr == nil && nn == 1 {
		} else {
			return "", errors.New("Unserialize Object len(Member) fail")
		}
		object := NewObject(cname)
		for i := 0; i < n; i++ {
			var s string
			fmt.Fscanf(r, "%1s", &s)
			key, kerr := unserializeKey(r, s == 's')
			if kerr != nil {
				return nil, kerr
			}
			val, verr := Unserialize(r)
			if verr != nil {
				return nil, verr
			}
			array.Set(key, val)
		}
		var s string
		fmt.Fscanf(r, "%1s", &s)
		if s == "}" {
			return array, nil
		}
	}
	return "", errors.New("Unserialize Array fail")
}

func Unserialize(r io.Reader) (ret PValue, err error) {
	var s string
	fmt.Fscanf(r, "%1s", &s)
	switch s {
	case "N":
		return unserializaNil(r)
	case "b":
		return unserializeBool(r)
	case "i":
		return unserializeLong(r)
	case "d":
		return unserializeDouble(r)
	case "s":
		return unserializeString(r)
	case "a":
		return unserializeArray(r)
	case "O":
		return unserializeObject(r)
	default:
		return nil, errors.New("Unknow value type")
	}
}

const (
	PublicVar      = 0
	ProtectedVar   = 1
	PrivateVar     = 2
	BasePrivateVar = 4
	endVarType     = 5
)

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
