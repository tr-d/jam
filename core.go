package jam

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	jmespath "github.com/jmespath/go-jmespath"
)

// Merge outputs the union of a and b with preference to b on matching keys.
func Merge(a, b interface{}) interface{} {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return b
	}
	switch b := b.(type) {
	case map[string]interface{}:
		a := a.(map[string]interface{})
		for k, v := range b {
			if u, ok := a[k]; ok {
				a[k] = Merge(u, v)
				continue
			}
			a[k] = v
		}
		return a
	case []interface{}:
		a := a.([]interface{})
		for k, v := range b {
			if k < len(a) {
				a[k] = Merge(a[k], v)
				continue
			}
			a = append(a, v)
		}
		return a
	default:
		return b
	}
}

// Diff ouputs c that satisfies Merge(a, c) == b. Transpose of Merge.
func Diff(a, b interface{}) interface{} {
	v, t := diff(a, b)
	if t {
		return nil
	}
	return v
}

// diff is the actual implementation of Diff which requires additional returns
// to work properly.
func diff(a, b interface{}) (o interface{}, t bool) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return b, false
	}
	if reflect.DeepEqual(a, b) {
		return b, true
	}
	switch b := b.(type) {
	case map[string]interface{}:
		a := a.(map[string]interface{})
		c := make(map[string]interface{}, len(b))
		for k, v := range b {
			if u, ok := a[k]; ok {
				if o, t := diff(u, v); !t {
					c[k] = o
				}
				continue
			}
			c[k] = v
		}
		return c, false
	case []interface{}:
		a := a.([]interface{})
		c := make([]interface{}, len(b))
		fill := false
		for i := len(b) - 1; i >= 0; i-- {
			if i >= len(a) {
				c[i] = b[i]
				fill = true
				continue
			}
			o, t := diff(a[i], b[i])
			switch {
			case !t:
				fill = true
				c[i] = o
			case fill:
				c[i] = a[i]
			default:
				c = c[:i]
			}
		}
		return c, !fill
	default:
		return b, false
	}
}

type filterer struct {
	i, r bool
	p    string
}

func (f filterer) filter(v interface{}, path string) (interface{}, bool) {
	// fmt.Println(path, "\t", v)
	if path == "" {
		if f.i {
			return nil, false
		}
		return v, true
	}
	if u, _, ok := nextValue(path); ok {
		t := reflect.DeepEqual(u, v)
		switch {
		case t && f.i:
			return nil, false
		case t:
			return v, true
		}
	}
	switch v := v.(type) {
	case map[string]interface{}:
		key, path, _ := nextKey(path)
		o := map[string]interface{}{}
		for k, v := range v {
			switch {
			case key == "*" || key == k:
				if tmp, ok := f.filter(v, path); ok {
					o[k] = tmp
				}
			case f.r:
				if tmp, ok := f.filter(v, f.p); ok {
					o[k] = tmp
				}
			case f.i:
				o[k] = v
			}
		}
		return o, f.i || len(o) > 0
	case []interface{}:
		lb, ub, path, ok := nextSlice(path, len(v))
		o := []interface{}{}
		for i := 0; i < len(v); i++ {
			switch {
			case f.r && i >= lb && i < ub:
				a, aok := f.filter(v[i], path)
				b, bok := f.filter(a, f.p)
				switch {
				case bok:
					o = append(o, b)
				case aok:
					o = append(o, a)
				}
			case i >= lb && i < ub:
				if tmp, ok := f.filter(v[i], path); ok {
					o = append(o, tmp)
				}
			case f.r && !ok:
				if tmp, ok := f.filter(v[i], f.p); ok {
					o = append(o, tmp)
				}
			case f.i:
				o = append(o, v)
			}
		}
		return o, f.i || len(o) > 0
	}
	if f.i {
		return v, true
	}
	return nil, false
}

// Filter filters a structure according to the path.
// Elements of v that do not match are removed.
// The path must match from the root of v.
func Filter(v interface{}, path string) interface{} {
	v, ok := filterer{}.filter(v, path)
	if !ok {
		return nil
	}
	return v
}

// FilterI filters a structure according to the path. Inverted.
// Elements of v that match are removed.
// The path must match from the root of v.
func FilterI(v interface{}, path string) interface{} {
	v, ok := filterer{i: true}.filter(v, path)
	if !ok {
		return nil
	}
	return v
}

// FilterR filters a structure according to the path. Recursive.
// Elements of v that do not match are removed.
// The path may match at any depth in v.
func FilterR(v interface{}, path string) interface{} {
	v, ok := filterer{r: true, p: path}.filter(v, path)
	if !ok {
		return nil
	}
	return v
}

// FilterIR filters a structure according to the path. Inverted and Recursive.
// Elements of v that match are removed.
// The path may match at any depth in v.
func FilterIR(v interface{}, path string) interface{} {
	v, ok := filterer{i: true, r: true, p: path}.filter(v, path)
	if !ok {
		return nil
	}
	return v
}

// Query applies a jmespath search to v.
func Query(v interface{}, s string) interface{} {
	v, _ = jmespath.Search(s, v)
	return v
}

// DropNil drops keys with value nil.
func DropNil(a interface{}) interface{} {
	switch a := a.(type) {
	case map[string]interface{}:
		b := make(map[string]interface{})
		for k, v := range a {
			_v := DropNil(v)
			if _v != nil {
				b[k] = _v
			}
		}
		return b
	case []interface{}:
		b := []interface{}{}
		for _, v := range a {
			_v := DropNil(v)
			if _v != nil {
				b = append(b, _v)
			}
		}
		return b
	default:
		return a
	}
}

// goStruct writes a go struct def to w
func goStruct(w io.Writer, v interface{}) {
	switch v := v.(type) {
	case map[string]interface{}:
		io.WriteString(w, "struct {\n")
		ks := make([]string, 0, len(v))
		for k, _ := range v {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		for _, k := range ks {
			// TODO: deal with an empty key
			if len(k) == 0 {
				continue
			}
			id := goCase(k)
			io.WriteString(w, id+" ")

			goStruct(w, v[k])

			if id != k {
				io.WriteString(w, " `json:\""+k+"\"`")
			}

			io.WriteString(w, "\n")
		}
		io.WriteString(w, "}")

	case []interface{}:
		io.WriteString(w, "[]")
		if len(v) == 0 {
			io.WriteString(w, "interface{}")
			return
		}
		t := reflect.TypeOf(v[0])
		var m interface{}
		for _, v := range v {
			if t != reflect.TypeOf(v) {
				io.WriteString(w, "interface{}")
				return
			}
			switch v.(type) {
			case map[string]interface{}:
				m = Merge(m, v)
			default:
			}
		}
		if m != nil {
			goStruct(w, m)
			return
		}
		goStruct(w, v[0])

	case nil:
		io.WriteString(w, "interface{}")
	default:
		fmt.Fprintf(w, "%T", v)
	}
}

// goCase turns a map key to go case
func goCase(s string) string {
	rs := []rune(s)
	id := ""
	first := true
	for i, r := range rs {
		switch {
		case unicode.IsLetter(r) && first:
			first = false
			id += string(unicode.ToUpper(r))
		case unicode.IsLetter(r):
			id += string(r)
		case unicode.IsDigit(r) && !first:
			id += string(r)
		case i < len(rs)-1 && unicode.IsLetter(rs[i+1]):
			rs[i+1] = unicode.ToUpper(rs[i+1])
		}
	}
	return id
}

var (
	nextKeyRe   = regexp.MustCompile(`^([^\.\[=]+)\.?`)
	nextSliceRe = regexp.MustCompile(`^\[(\d*)(?:(:?)(\d*))?\]\.?`)
	// nextValueRe = regexp.MustCompile(`^\[==([^\]]+)\]\.?`)
	nextValueRe = regexp.MustCompile(`^==(.+)$`)
)

func nextKey(path string) (string, string, bool) {
	ms := nextKeyRe.FindStringSubmatch(path)
	if len(ms) == 0 {
		return "", path, false
	}
	return ms[1], path[len(ms[0]):], true
}

func nextSlice(path string, length int) (int, int, string, bool) {
	ms := nextSliceRe.FindStringSubmatch(path)
	if len(ms) == 0 {
		return 0, 0, path, false
	}

	var lb, ub int
	switch {
	case ms[3] == "" && (ms[1] == "" || ms[2] == ":"):
		lb, _ = strconv.Atoi(ms[1])
		ub = length
	case ms[3] == "":
		lb, _ = strconv.Atoi(ms[1])
		ub = lb + 1
	default:
		lb, _ = strconv.Atoi(ms[1])
		ub, _ = strconv.Atoi(ms[3])
	}
	return lb, ub, path[len(ms[0]):], true
}

func nextValue(path string) (interface{}, string, bool) {
	var v interface{}
	ms := nextValueRe.FindStringSubmatch(path)
	if len(ms) == 0 {
		return v, path, false
	}
	err := NewDecoder(strings.NewReader(ms[1])).Decode(&v)
	return v, path[len(ms[0]):], err == nil
}
