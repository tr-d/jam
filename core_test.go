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
		j := &Jam{}
		j.Merge(s.a)
		j.Merge(s.b)
		if !reflect.DeepEqual(j.Value(0), s.x) {
			t.Errorf("Expected %v, got %v", s.x, j.Value(0))
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
		j := &Jam{}
		j.Diff(s.a)
		j.Diff(s.b)
		if !reflect.DeepEqual(j.Value(0), s.x) {
			t.Errorf("Expected %v, got %v", s.x, j.Value(0))
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
		j := NewJam(s.d)
		j.Filter(s.q)
		if !reflect.DeepEqual(j.Value(0), s.x) {
			t.Errorf("for %q, expected %v, got %v", s.q, s.x, j.Value(0))
		}
	}
}

func TestFilterI(t *testing.T) {
	var ss = []struct {
		q    string
		d, x interface{}
	}{
		{"", true, nil},
		{"nah", true, true},
		{"nah", _s{}, _s{}},
		{"[]", _m{}, _m{}},
		{"[]", _s{"a", "b", "c"}, _s{}},
		{"[0]", _s{"a", "b", "c"}, _s{"b", "c"}},
		{"[1]", _s{"a", "b", "c"}, _s{"a", "c"}},
		{"foo", _m{"foo": true, "baz": true}, _m{"baz": true}},
		{"*", _m{"foo": true, "baz": true}, _m{}},
		{"foo[]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{}}},
		{"foo[:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{}}},
		{"foo[0]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{2, 3}}},
		{"foo[2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2}}},
		{"foo[1:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1}}},
		{"foo[:2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{3}}},
		{"[].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"y": "y"}, _m{"y": "x"}}},
		{"[1:2].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"x": 1, "y": "y"}, _m{"y": "x"}}},
		{
			"[][]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{}, _s{}, _s{}},
		},
		{
			"[1:][:2]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{1, 2, 3}, _s{6}, _s{9}},
		},
		{
			"foo.foo",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{"baz": true}},
		},
		{
			"foo.*",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{}},
		},
		{"==1", float64(1), nil},
		{"[]==blep", _s{"blep", "mlem"}, _s{"mlem"}},
		{"*==mlem", _m{"0": "blep", "1": "mlem"}, _m{"0": "blep"}},
	}

	for _, s := range ss {
		j := NewJam(s.d)
		j.FilterI(s.q)
		if !reflect.DeepEqual(j.Value(0), s.x) {
			t.Errorf("for %q, expected %v, got %v", s.q, s.x, j.Value(0))
		}
	}
}

func TestFilterR(t *testing.T) {
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
		j := NewJam(s.d)
		j.FilterR(s.q)
		if !reflect.DeepEqual(j.Value(0), s.x) {
			t.Errorf("for %q, expected %v, got %v", s.q, s.x, j.Value(0))
		}
	}
}

func TestFilterIR(t *testing.T) {
	var ss = []struct {
		q    string
		d, x interface{}
	}{
		{"", true, nil},
		{"nah", true, true},
		{"[]", _s{"a", "b", "c"}, _s{}},
		{"[0]", _s{"a", "b", "c"}, _s{"b", "c"}},
		{"[1]", _s{"a", "b", "c"}, _s{"a", "c"}},
		{"foo", _m{"foo": true, "baz": true}, _m{"baz": true}},
		{"*", _m{"foo": true, "baz": true}, _m{}},
		{"foo[]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{}}},
		{"foo[:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{}}},
		{"foo[0]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{2, 3}}},
		{"foo[2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1, 2}}},
		{"foo[1:]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{1}}},
		{"foo[:2]", _m{"foo": _s{1, 2, 3}}, _m{"foo": _s{3}}},
		{"[].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"y": "y"}, _m{"y": "x"}}},
		{"[1:2].x", _s{_m{"x": 1, "y": "y"}, _m{"x": 2, "y": "x"}}, _s{_m{"x": 1, "y": "y"}, _m{"y": "x"}}},
		{
			"[][]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{}, _s{}, _s{}},
		},
		{
			"[1:][:2]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{1, 2, 3}, _s{6}, _s{9}},
		},
		{
			"foo.foo",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{"baz": true}},
		},
		{
			"foo.*",
			_m{"foo": _m{"foo": true, "baz": true}},
			_m{"foo": _m{}},
		},
		{
			"blep",
			_m{"foo": _m{"blep": true, "baz": true}},
			_m{"foo": _m{"baz": true}},
		},
		{
			"mlem",
			_s{_m{"mlem": true, "baz": true}},
			_s{_m{"baz": true}},
		},
		{
			"[]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{},
		},
		{
			"[1:]",
			_s{_s{1, 2, 3}, _s{4, 5, 6}, _s{7, 8, 9}},
			_s{_s{1, 2, 3}},
		},
		{"==blep", _s{"blep", "mlem"}, _s{"mlem"}},
		{"==mlem", _m{"0": "blep", "1": "mlem"}, _m{"0": "blep"}},
	}

	for _, s := range ss {
		j := NewJam(s.d)
		j.FilterIR(s.q)
		if !reflect.DeepEqual(j.Value(0), s.x) {
			t.Errorf("for %q, expected %v, got %v", s.q, s.x, j.Value(0))
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
