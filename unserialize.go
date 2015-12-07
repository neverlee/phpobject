package phpobject

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func unserializaNil(r io.Reader) (ret PValue, err error) {
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
		buf := make([]byte, l+2)
		if _, berr := io.ReadFull(r, buf); berr == nil && buf[l] == '"' && buf[l+1] == ';' {
			return PString(buf[:l]), nil
		}
	}
	return "", errors.New("Unserialize String fail")
}

func unserializeKey(r io.Reader, isstr bool) (ret string, err error) {
	if isstr {
		if s, serr := unserializeString(r); serr == nil {
			return string(s), nil
		}
	} else {
		if l, lerr := unserializeLong(r); lerr == nil {
			return strconv.Itoa(int(l)), lerr
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
			key, kerr := unserializeKey(r, s == "s")
			if kerr != nil {
				return nil, kerr
			}
			val, verr := unserializeValue(r)
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
	return nil, errors.New("Unserialize Array fail")
}

func unserializeVarname(r io.Reader) (clsname, varname string, err error) {
	var s string
	fmt.Fscanf(r, "%1s", &s)
	if cname, cerr := unserializeString(r); cerr == nil {
		slist := strings.Split(string(cname), "\x00")
		if len(slist) == 2 {
			return slist[0], slist[1], nil
		} else if len(slist) == 1 {
			return "", slist[0], nil
		} else {
			return "", "", errors.New("Unserialize Varname fail")
		}
	} else {
		return "", "", errors.New("Unserialize Varname fail")
	}
}

func unserializeObject(r io.Reader) (ret *PObject, err error) {
	var cl int
	if ln, lerr := fmt.Fscanf(r, ":%d:\"", &cl); lerr == nil && ln == 1 {
		cnbuf := make([]byte, cl+1)
		var cname string
		if _, cerr := io.ReadFull(r, cnbuf); cerr == nil && cnbuf[cl] == '"' {
			cname = string(cnbuf[:cl])
		} else {
			return nil, errors.New("Unserialize Object Name fail")
		}
		var n int
		if nn, nerr := fmt.Fscanf(r, ":%d:{", &n); nerr == nil && nn == 1 {
		} else {
			return nil, errors.New("Unserialize Object len(Member) fail")
		}
		object := NewObject(cname)
		for i := 0; i < n; i++ {
			clsname, varname, verr := unserializeVarname(r)
			if verr != nil {
				return nil, verr
			}
			val, pverr := unserializeValue(r)
			if pverr != nil {
				return nil, errors.New("Unserialize Object len(Member) fail")
			}
			if object.Set(clsname, varname, val) != nil {
				return nil, errors.New("Unserialize Object set fail")
			}
		}
		var s string
		fmt.Fscanf(r, "%1s", &s)
		if s == "}" {
			return object, nil
		}
	}
	return nil, errors.New("Unserialize Object fail")
}

func unserializeValue(r io.Reader) (ret PValue, err error) {
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

func Unserialize(r io.Reader) (ret PValue, err error) {
	return unserializeValue(r)
}
