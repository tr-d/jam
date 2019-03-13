// Package jam. Decode yaml, json or toml. Encode yaml, json, toml, go syntax and
// go struct defs.  Merge, Diff, Filter, Query things. Struct tags.
//
// If structured data is scones and you are clotted cream, this is jam.
package jam

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"reflect"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/ghodss/yaml"
	jmespath "github.com/jmespath/go-jmespath"
)

// Decoder reads yaml, json, or toml from a reader, "jam" struct tags are
// evaluated as jmespath expressions.
type Decoder struct {
	r    io.Reader
	jd   *json.Decoder
	once bool
}

// NewDecoder creates a Decoder for r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode json, yaml, or toml from the reader and store the result in the value
// pointed to by v. Struct tags labelled "jam" are evaluated as jmespath
// expressions.
func (d *Decoder) Decode(v interface{}) error {
	defer func() { d.once = true }()
	var (
		err error
		u   interface{}
	)
	switch {
	case d.jd != nil:
		if !d.jd.More() {
			return ErrNoMore{}
		}
		if err := d.jd.Decode(&u); err != nil {
			return err
		}
	default:
		b, err := ioutil.ReadAll(d.r)
		if err != nil {
			return err
		}
		if d.once && len(b) == 0 {
			return ErrNoMore{}
		}
		a := analyze(b)
		switch a.lang {
		case lToml:
			_, err = toml.Decode(string(b), &u)
			if err != nil {
				return err
			}
		default:
			if !d.once {
				// this step protects json from yaml specific errs
				jd := json.NewDecoder(bytes.NewReader(b))
				if err = jd.Decode(&u); err == nil {
					d.jd = jd
					break
				}
			}
			if len(a.errs) > 0 {
				return a.nerrs(6)
			}
			b = bytes.TrimPrefix(b, []byte("---\n"))
			// \n---\n can only be a yaml document separator
			// this is a safe split, apparently
			bs := bytes.SplitN(b, []byte("\n---\n"), 2)
			err = yaml.Unmarshal(bs[0], &u)
			if len(bs) > 1 {
				d.r = bytes.NewReader(bs[1])
			}
			if err != nil {
				return err
			}
		}
	}

	u, err = remap(u, reflect.TypeOf(v))
	if err != nil {
		return err
	}

	bb := bytes.NewBuffer([]byte{})
	if err := NewEncoder(bb).AsJson().Encode(u); err != nil {
		return err
	}

	return json.NewDecoder(bb).Decode(v)
}

// Encoder writes yaml, json, toml, go syntax, or go struct definition to a
// writer.  The behaviour depends on the underlying function, which may be set
// using the AsYaml, AsJson, AsToml, AsGo, and AsStruct methods. The default is
// yaml.
type Encoder struct {
	w      io.Writer
	encode func(w io.Writer, v interface{}) error
}

// NewEncoder creates an Encoder set to encode as yaml.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, encode: asYaml}
}

// AsGo creates a copy of this Encoder set to create go syntax.
func (e *Encoder) AsGo() *Encoder {
	return &Encoder{w: e.w, encode: asGo}
}

// AsStruct creates a copy of this Encoder set to create go struct definitions.
func (e *Encoder) AsStruct() *Encoder {
	return &Encoder{w: e.w, encode: asStruct}
}

// AsYaml creates a copy of this Encoder set to encode as yaml.
func (e *Encoder) AsYaml() *Encoder {
	return &Encoder{w: e.w, encode: asYaml}
}

// AsJson creates a copy of this Encoder set to encode as json.
func (e *Encoder) AsJson() *Encoder {
	return &Encoder{w: e.w, encode: asJson}
}

// AsToml creates a copy of this Encoder set to encode as toml.
func (e *Encoder) AsToml() *Encoder {
	return &Encoder{w: e.w, encode: asToml}
}

// Encode writes to the underlying writer.  The behaviour depends on the
// underlying function, which may be set using the AsYaml, AsJson, AsToml, AsGo,
// AsStruct methods. The default is yaml.
func (e *Encoder) Encode(v interface{}) error {
	return e.encode(e.w, v)
}

type ErrNoMore struct{}

func (_ ErrNoMore) Error() string {
	return "no more to decode"
}

func IsNoMore(e error) bool {
	_, ok := e.(ErrNoMore)
	return ok
}

// Jam accumulates operations on a data tree.
type Jam struct {
	vs []interface{}
}

func (j *Jam) atLeast(length int) {
	for i := len(j.vs); i < length; i++ {
		j.vs = append(j.vs, interface{}(nil))
	}
}

// NewJam creates a Jam.
func NewJam(vs ...interface{}) *Jam {
	return &Jam{vs}
}

func (j *Jam) Diff(vs ...interface{}) {
	j.atLeast(len(vs))
	for i, v := range vs {
		j.vs[i] = Diff(j.vs[i], v)
	}
}

func (j *Jam) Merge(vs ...interface{}) {
	j.atLeast(len(vs))
	for i, v := range vs {
		j.vs[i] = Merge(j.vs[i], v)
	}
}

func (j *Jam) Exec(dst io.Writer, src io.Reader) error {
	var bb bytes.Buffer
	_, err := io.Copy(&bb, src)
	if err != nil {
		return err
	}
	t, err := template.New("").Parse(bb.String())
	if err != nil {
		return err
	}
	for _, v := range j.vs {
		if err := t.Execute(dst, v); err != nil {
			return err
		}
	}
	return nil
}

// Query applies the Query function to the Jam's value.
func (j *Jam) Query(q string) {
	for i := range j.vs {
		j.vs[i] = Query(j.vs[i], q)
	}
}

// Filter applies the Filter function to the Jam's value.
func (j *Jam) Filter(q string) {
	for i := range j.vs {
		j.vs[i] = Filter(j.vs[i], q)
	}
}

// FilterI applies the FilterI function to the Jam's value.
func (j *Jam) FilterI(q string) {
	for i := range j.vs {
		j.vs[i] = FilterI(j.vs[i], q)
	}
}

// FilterR applies the FilterR function to the Jam's value.
func (j *Jam) FilterR(q string) {
	for i := range j.vs {
		j.vs[i] = FilterR(j.vs[i], q)
	}
}

// FilterIR applies the FilterIR function to the Jam's value.
func (j *Jam) FilterIR(q string) {
	for i := range j.vs {
		j.vs[i] = FilterIR(j.vs[i], q)
	}
}

// Value returns the Jam's value.
func (j *Jam) Value(i int) interface{} {
	if i >= len(j.vs) {
		return nil
	}
	return j.vs[i]
}

func (j *Jam) Values() []interface{} {
	return j.vs
}

// analysis is the result of lang and error analysis.
type analysis struct {
	lang lang
	errs []error
}

// analyze creates a new analysis.
func analyze(b []byte) analysis {
	a := &analysis{errs: []error{}}
	bloop(b, a.hasTab(), a.hasTag(), a.isLang())
	return *a
}

// nerrs returns up to n errors from the analysis on separate lines
func (a *analysis) nerrs(n int) error {
	lim := len(a.errs)
	if lim > n {
		lim = n
	}
	s := ""
	for i := 0; i < lim; i++ {
		s += a.errs[i].Error() + "\n"
	}
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	if len(a.errs) > lim {
		s += fmt.Sprintf("\n%d more errors", len(a.errs)-lim)
	}
	return errors.New(s)
}

// hasTab yields a function to find tabs in yaml indent
func (a *analysis) hasTab() func(byte, ref) {
	danger := true
	return func(c byte, r ref) {
		switch {
		case c == '\n':
			danger = true
		case !danger:

		case c == ' ':
		case c == '\t':
			a.errs = append(a.errs, &tabErr{r: r})
		default:
			danger = false
		}
	}
}

// hasTag yields a function to find tags in yaml
func (a *analysis) hasTag() func(byte, ref) {
	warm, hot := false, true
	return func(c byte, r ref) {
		switch {
		case (c == ':' || c == '[' || c == '{' || c == ','):
			warm, hot = true, false
		case c == '\n':
			hot = true
		case !(warm || hot):

		case c == ' ':
			hot = true
		case !hot:

		case c == '!':
			a.errs = append(a.errs, &tagErr{r: r})
		default:
			warm, hot = false, false
		}
	}
}

// isLang yields a function to determine the language of a sample
func (a *analysis) isLang() func(byte, ref) {
	var hot bool
	return func(c byte, r ref) {
		switch {
		case c == ':':
			hot = true
		case hot && c == ' ':
			a.lang = lYaml
		case hot:
			hot = false
		case a.lang == lYaml:
		case c == '=':
			a.lang = lToml
		}
	}
}

type lang int

const (
	lUnknown lang = iota
	lYaml
	lJson
	lToml
)

// ref is a line and column number
type ref struct {
	l, c int
}

// tabErr is an error for tabs found in yaml indent
type tabErr struct{ r ref }

func (e *tabErr) Error() string {
	return fmt.Sprintf("%d:%d: yaml: tab indents are not valid", e.r.l, e.r.c)
}

// tagErr is an error for tags found in yaml
type tagErr struct{ r ref }

func (e *tagErr) Error() string {
	return fmt.Sprintf("%d:%d: yaml: tags are not supported", e.r.l, e.r.c)
}

// asGo writes formatted go syntax to w
func asGo(w io.Writer, v interface{}) error {
	s := fmt.Sprintf("%+#v\n", v)
	t := ""
	var escape, bquote, squote, dquote bool
	for i, c := range s {
		t += string(c)
		switch {
		case escape:
			escape = false
		case c == '\\':
			escape = true
		case c == '`' && !dquote && !squote:
			bquote = !bquote
		case c == '"' && !bquote && !squote:
			dquote = !dquote
		case c == '\'' && !bquote && !dquote:
			squote = !squote
		}

		switch {
		case bquote || squote || dquote || escape:
		case c == ',' || c == ';':
			t += "\n"
		case i >= len(s)-1:
		case c == '{' && s[i+1] != '}':
			t += "\n"
		case c != '{' && s[i+1] == '}' && i < len(s)-2 && s[i+2] == '{':
			t += "\n"
		case c != '{' && s[i+1] == '}':
			t += ",\n"
		}
	}
	b, err := format.Source([]byte(t))
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

// asJson writes to json to w.
func asJson(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}

// asStruct writes a formatted go struct definition to w.
func asStruct(w io.Writer, v interface{}) error {
	bb := bytes.NewBuffer([]byte{})
	io.WriteString(bb, "type T ")
	goStruct(bb, v)
	io.WriteString(bb, "\n")

	b, err := format.Source(bb.Bytes())
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

// asToml writes toml to w.
func asToml(w io.Writer, y interface{}) error {
	return toml.NewEncoder(w).Encode(y)
}

// asYaml writes yaml to w.
func asYaml(w io.Writer, v interface{}) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	ds := []byte("---\n")
	if !bytes.HasPrefix(b, ds) {
		b = append(ds, b...)
	}
	_, err = w.Write(b)
	return err
}

// bloop calls funcs with each byte of b that is unquoted and unescaped,
// and a reference to the line and column number
func bloop(b []byte, funcs ...func(byte, ref)) error {
	var escape, squote, dquote bool
	line, col := 1, 0
	for _, c := range b {
		switch {
		case c == '\n':
			line++
			col = 0
		default:
			col++
		}

		switch {
		case escape:
			escape = false
		case c == '\\':
			escape = true

		case !dquote && c == '\'':
			squote = !squote
		case !squote && c == '"':
			dquote = !dquote
		case squote || dquote:

		default:
			for _, fn := range funcs {
				fn(c, ref{l: line, c: col})
			}
		}
	}
	return nil
}

// remap remaps data onto a type using "jam" struct tags. The result will be json decodable
// into t. Struct tags are evaluated as jmespath expressions.
func remap(data interface{}, t reflect.Type) (interface{}, error) {
	if data == nil || t == nil {
		return nil, nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		m := map[string]interface{}{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			x, err := jmespath.Search(f.Tag.Get("jam"), data)
			if err != nil {
				return data, fmt.Errorf("failed to unmarshal into %s: %s", f.Name, err)
			}
			x, err = remap(x, f.Type)
			if err != nil {
				return data, err
			}
			n := f.Tag.Get("json")
			if n == "" {
				n = f.Name
			}
			m[n] = x
		}
		data = m
	}
	return data, nil
}
