package jam

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type _m = map[string]interface{}
type _s = []interface{}

func TestYJ(t *testing.T) {
	var ss = []interface{}{
		_m{"b": true},
		_m{"f": 13.37},
		_m{"s": "foo baz"},
		_m{"a": _s{"a", "b", "x"}},
		_m{"foo \"= baz": true},
	}
	for _, s := range ss {
		testYaml(t, s)
		testJson(t, s)
		testToml(t, s)
	}
}

func testYaml(t *testing.T, u interface{}) {
	bb := bytes.NewBuffer([]byte{})
	if err := NewEncoder(bb).Encode(u); err != nil {
		t.Error(err)
	}
	var v interface{}
	if err := NewDecoder(bb).Decode(&v); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(u, v) {
		t.Errorf("yaml: expected %#v, got %#v", u, v)
	}
}

func testJson(t *testing.T, u interface{}) {
	bb := bytes.NewBuffer([]byte{})
	if err := NewEncoder(bb).AsJson().Encode(u); err != nil {
		t.Error(err)
	}
	var v interface{}
	if err := NewDecoder(bb).Decode(&v); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(u, v) {
		t.Errorf("json: expected %#v, got %#v", u, v)
	}
}

func testToml(t *testing.T, u interface{}) {
	bb := bytes.NewBuffer([]byte{})
	if err := NewEncoder(bb).AsToml().Encode(u); err != nil {
		t.Error(err)
	}
	var v interface{}
	if err := NewDecoder(bb).Decode(&v); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(u, v) {
		t.Errorf("toml: expected %#v, got %#v", u, v)
	}
}

func TestDecodeFail(t *testing.T) {
	ss := []string{
		"=",
		"foo:\n\tbaz: true",
	}
	var v interface{}
	for i, s := range ss {
		if err := NewDecoder(strings.NewReader(s)).Decode(&v); err == nil {
			t.Errorf("Expected error got none [%d].", i)
		}
	}
}

func TestDecodePass(t *testing.T) {
	ss := []string{
		"",
		"foo:\n  baz: true",
		"{\n\t\"x\":1\n}",
	}
	var v interface{}
	for i, s := range ss {
		if err := NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
			t.Errorf("Got error [%d]: %s", i, err)
		}
	}
}

func TestStruct(t *testing.T) {
	var ss = []struct {
		i interface{}
		x string
	}{
		{0, "type T int\n"},
		{"", "type T string\n"},
		{0.5, "type T float64\n"},
		{nil, "type T interface{}\n"},
		{struct{}{}, "type T struct{}\n"},
		{[]int{}, "type T []int\n"},
		{[]string{}, "type T []string\n"},
		{[]float64{}, "type T []float64\n"},
		{[]struct{}{}, "type T []struct{}\n"},
		{_s{}, "type T []interface{}\n"},
		{_s{1, 2}, "type T []int\n"},
		{_s{1, true}, "type T []interface{}\n"},
		{_s{"a", "b"}, "type T []string\n"},
		{_s{"", true}, "type T []interface{}\n"},
		{_m{"x": 0}, "type T struct {\n\tX int `json:\"x\"`\n}\n"},
		{_m{"x-x": 0}, "type T struct {\n\tXX int `json:\"x-x\"`\n}\n"},
		{_m{"x0": 0}, "type T struct {\n\tX0 int `json:\"x0\"`\n}\n"},
		{
			_s{_m{"A": 0}, _m{"B": ""}, _m{"C": true}},
			"type T []struct {\n\tA int\n\tB string\n\tC bool\n}\n",
		},
	}
	for _, s := range ss {
		bb := bytes.NewBuffer([]byte{})
		NewEncoder(bb).AsStruct().Encode(s.i)
		if s.x != bb.String() {
			t.Errorf("struct: expected\n%s\ngot\n%s", s.x, bb.String())
		}
	}
}

func TestADotYaml(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "a.yml"))
	if err != nil {
		t.Error(err)
	}
	var v interface{}
	if err := NewDecoder(f).Decode(&v); err != nil {
		t.Error(err)
	}
	f.Close()

	bb := bytes.NewBuffer([]byte{})
	e := NewEncoder(bb)
	var ss = []struct {
		f string
		e *Encoder
	}{
		{"a.yml", e},
		{"a.json", e.AsJson()},
		{"a.toml", e.AsToml()},
	}

	for _, s := range ss {
		bb.Reset()
		if err := s.e.Encode(v); err != nil {
			t.Error(err)
		}

		var r interface{}
		if err := NewDecoder(bb).Decode(&r); err != nil {
			t.Error(err)
		}
		f, err := os.Open(filepath.Join("testdata", s.f))
		if err != nil {
			t.Error(err)
		}

		var x interface{}
		if err := NewDecoder(f).Decode(&x); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(x, r) {
			t.Error("does not match: " + s.f)
		}
	}

	bb.Reset()
	if err := e.AsStruct().Encode(v); err != nil {
		t.Error(err)
	}
	x, err := ioutil.ReadFile(filepath.Join("testdata", "a.struct"))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(x, bb.Bytes()) {
		t.Error("does not match: a.struct")
	}
}

func TestLang(t *testing.T) {
	var ss = []struct {
		i string
		x lang
	}{
		{"foo: baz", lYaml},
		{"=foo: baz", lYaml},
		{"foo: =baz", lYaml},
		{"foo = \"baz\"", lToml},
		{"foo = 1979-05-27T07:32:00Z", lToml},
		{"foo = 1979-05-27T00:32:00-07:00", lToml},
		{"foo = 1979-05-27T00:32:00.999999-07:00", lToml},
		{"foo = 1979-05-27 07:32:00Z", lToml},
	}
	for _, s := range ss {
		if o := analyze([]byte(s.i)).lang; o != s.x {
			t.Errorf("%v: expecting %v, got %v", s.i, s.x, o)
		}
	}
}

func TestTag(t *testing.T) {
	ss := []string{
		"!", " !", "  !", "[ ! ]", "[    ! ]", "[ '', ! ]", "foo: !baz", "!foo: baz",
	}
	for _, s := range ss {
		a := analyze([]byte(s))
		if len(a.errs) != 1 {
			t.Errorf("%v: expecting one err, got %d", s, len(a.errs))
		}
		if _, ok := a.errs[0].(*tagErr); !ok {
			t.Errorf("%v: error is not a tagErr", s)
		}
	}

	ss = []string{
		"'!'", "foo: baz", "foo: baz !baz", "foo !baz: baz", "foo!baz: baz", "foo: \"!baz\"",
	}
	for _, s := range ss {
		a := analyze([]byte(s))
		if len(a.errs) > 0 {
			t.Errorf("%v: expecting no errors, got %d", s, len(a.errs))
		}
	}
}

func TestTab(t *testing.T) {
	ss := []string{
		"\t", "\tx", "\t x", " \tx", "  \t x", "\n\t", "\n\tx", "\n \tx", "\n  \t x",
	}
	for _, s := range ss {
		a := analyze([]byte(s))
		if len(a.errs) != 1 {
			t.Errorf("%v: expecting one err, got %d", s, len(a.errs))
		}
		if _, ok := a.errs[0].(*tabErr); !ok {
			t.Errorf("%v: error is not a tabErr", s)
		}
	}

	ss = []string{"x:\ty", "foo: baz \t"}
	for _, s := range ss {
		a := analyze([]byte(s))
		if len(a.errs) != 0 {
			t.Errorf("%v: expecting no errors, got %d", s, len(a.errs))
		}
	}
}
