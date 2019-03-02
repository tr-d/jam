package jam

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	var ss = []struct {
		a, b, x interface{}
	}{
		{0, false, false},
		{false, true, true},
		{_m{"foo": true}, _m{"foo": false}, _m{"foo": false}},
		{_m{"foo": true}, _m{"baz": true}, _m{"foo": true, "baz": true}},
		{_s{true, "foo"}, _s{false}, _s{false, "foo"}},
		{_s{}, _s{true}, _s{true}},
	}

	for _, s := range ss {
		o := Merge(s.a, s.b)
		if !reflect.DeepEqual(o, s.x) {
			t.Errorf("Expected %v, got %v", s.x, o)
		}
	}
}

func TestDiff(t *testing.T) {
	var ss = []struct {
		a, b, x interface{}
	}{
		{0, 0, nil},
		{_s{0, 1, 2}, _s{0, 1, 2, 3}, _s{0, 1, 2, 3}},
		{_s{0, 1, 2}, _s{2, 3, 2}, _s{2, 3}},
		{_s{0, 1, 2}, _s{0, 3, 2}, _s{0, 3}},
		{0, false, false},
		{false, true, true},
		{_m{"foo": true}, _m{"foo": false}, _m{"foo": false}},
		{_m{"foo": true}, _m{"baz": true}, _m{"baz": true}},
		{_s{true, "foo"}, _s{false}, _s{false}},
		{_s{}, _s{true}, _s{true}},
	}

	for _, s := range ss {
		o := Diff(s.a, s.b)
		if !reflect.DeepEqual(o, s.x) {
			t.Errorf("Expected %v, got %v", s.x, o)
		}
	}
}

func TestFilter(t *testing.T) {
	var ss = []struct {
		q    string
		d, x interface{}
	}{
		{"", true, true},
		{"nah", true, nil},
		{"nah", _s{}, nil},
		{"[]", _m{}, nil},
		{"[]", _s{"a", "b", "c"}, _s{"a", "b", "c"}},
		{"[0]", _s{"a", "b", "c"}, _s{"a"}},
		{"[1]", _s{"a", "b", "c"}, _s{"b"}},
		{"foo", _m{"foo": true, "baz": true}, _m{"foo": true}},
		{"*", _m{"foo": true, "baz": true}, _m{"foo": true, "baz": true}},
		{"foo[]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2, 3}}},
		{"foo[:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2, 3}}},
		{"foo[0]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1}}},
		{"foo[2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{3}}},
		{"foo[1:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{2, 3}}},
		{"foo[:2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2}}},
		{"[].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"x": 1}, _m{"x": 2}}},
		{"[1:2].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"x": 2}}},
		{
			"[][]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
		},
		{
			"[1:][:2]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{4, 5}, _s{7, 8}},
		},
		{
			"foo.foo",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{"foo": true}},
		},
		{
			"foo.*",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{"foo": true, "baz": true}},
		},
		{"==1", float64(1), float64(1)},
		{"[]==blep", _s{"blep", "mlem"}, _s{"blep"}},
		{"*==mlem", _m{"0": "blep", "1": "mlem"}, _m{"1": "mlem"}},
	}

	for _, s := range ss {
		o := Filter(s.d, s.q)
		if !reflect.DeepEqual(o, s.x) {
			t.Errorf("for %q, expected %v, got %v", s.q, s.x, o)
		}
	}
}

func TestRFilter(t *testing.T) {
	var ss = []struct {
		q    string
		d, x interface{}
	}{
		{"", true, true},
		{"nah", true, nil},
		{"[]", _s{"a", "b", "c"}, _s{"a", "b", "c"}},
		{"[0]", _s{"a", "b", "c"}, _s{"a"}},
		{"[1]", _s{"a", "b", "c"}, _s{"b"}},
		{"foo", _m{"foo": true, "baz": true}, _m{"foo": true}},
		{"*", _m{"foo": true, "baz": true}, _m{"foo": true, "baz": true}},
		{"foo[]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2, 3}}},
		{"foo[:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2, 3}}},
		{"foo[0]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1}}},
		{"foo[2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{3}}},
		{"foo[1:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{2, 3}}},
		{"foo[:2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2}}},
		{"[].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"x": 1}, _m{"x": 2}}},
		{"[1:2].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"x": 2}}},
		{
			"[][]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
		},
		{
			"[1:][:2]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{4, 5}, _s{7, 8}},
		},
		{
			"foo.foo",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{"foo": true}},
		},
		{
			"foo.*",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{"foo": true, "baz": true}},
		},
		{
			"blep",
			_m{"foo": _m{"blep": true, "baz": true}},
			_m{"foo": _m{"blep": true}},
		},
		{
			"mlem",
			_s{_m{"mlem": true, "baz": true}},
			_s{_m{"mlem": true}},
		},
		{
			"[]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
		},
		{
			"[1:]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{5, 6}, _s{8, 9}},
		},
		{"==blep", _s{"blep", "mlem"}, _s{"blep"}},
		{"==mlem", _m{"0": "blep", "1": "mlem"}, _m{"1": "mlem"}},
	}

	for _, s := range ss {
		o := FilterR(s.d, s.q)
		if !reflect.DeepEqual(o, s.x) {
			t.Errorf("for %q, expected %v, got %v", s.q, s.x, o)
		}
	}
}

func TestDropNil(t *testing.T) {
	var ss = []struct {
		a, x interface{}
	}{
		{0, 0},
		{false, false},
		{nil, nil},
		{_m{"foo": true}, _m{"foo": true}},
		{_m{"foo": nil}, _m{}},
		{_s{nil}, _s{}},
		{_m{"foo": true, "baz": nil}, _m{"foo": true}},
		{_s{true, nil}, _s{true}},
	}

	for _, s := range ss {
		o := FilterIR(s.a, "==null")
		if !reflect.DeepEqual(o, s.x) {
			t.Errorf("expected %v, got %v", s.x, o)
		}
	}
}
