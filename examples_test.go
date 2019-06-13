package jam_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/tr-d/jam"
)

func ExampleEncodeStructAsGo() {
	x := struct {
		A, B int
	}{1, 2}
	jam.NewEncoder(os.Stdout).AsGo().Encode(x)
	// Output:
	// struct {
	// 	A int
	// 	B int
	// }{
	// 	A: 1,
	// 	B: 2,
	// }
}

func ExampleEncodeStructAsJson() {
	x := struct {
		A, B int
	}{1, 2}
	jam.NewEncoder(os.Stdout).AsJson().Encode(x)
	// Output: {"A":1,"B":2}
}

func ExampleEncodeStructAsStruct() {
	x := struct {
		A, B int
	}{1, 2}
	jam.NewEncoder(os.Stdout).AsStruct().Encode(x)
	// Output:
	// type T struct {
	// 	A int
	// 	B int
	// }
}

func ExampleEncodeStructAsToml() {
	x := struct {
		A, B int
	}{1, 2}
	jam.NewEncoder(os.Stdout).AsToml().Encode(x)
	// Output:
	// A = 1
	// B = 2
}

func ExampleEncodeStructAsYaml() {
	x := struct {
		A, B int
	}{1, 2}
	jam.NewEncoder(os.Stdout).Encode(x)
	// Output:
	// ---
	// A: 1
	// B: 2
}

func ExampleStructTags() {
	s := `{"a":{"a":{"a":"blep"},"b":{"a":"mlem"}}}`
	v := struct {
		A string `jam:"a.a.a"`
		B string `jam:"a.b.a"`
	}{}
	jam.NewDecoder(strings.NewReader(s)).Decode(&v)
	fmt.Println(v)
	// Output: {blep mlem}
}

func ExampleStructTagsPlus() {
	s := `[{"a":"blep","b":1},{"a":"mlem","b":-1}]`
	v := struct {
		As []string `jam:"[].a" json:"as"`
		Bs []int    `jam:"[].b" json:"bs"`
	}{}
	jam.NewDecoder(strings.NewReader(s)).Decode(&v)
	jam.NewEncoder(os.Stdout).AsJson().Encode(v)
	// Output: {"as":["blep","mlem"],"bs":[1,-1]}
}

func ExampleDecoderMerge() {
	a := strings.NewReader(`a: blep`)
	b := strings.NewReader(`{"b":"blep"}`)
	c := strings.NewReader(`b = "mlem"`)
	v := struct {
		A string `json:"a"`
		B string `json:"b"`
	}{}
	jam.NewDecoder(a, b, c).Decode(&v)
	fmt.Printf("%+v\n", v)
	// Output: {A:blep B:mlem}
}

func ExampleFileDecoderMerge() {
	ss := []string{
		"testdata/standard.yml",
		"testdata/moar.json",
		"testdata/extra.toml",
	}
	v := struct {
		Thing  string `jam:"config.thing"`
		Other  string `jam:"config.otherThing"`
		Third  string `jam:"config.thirdThing"`
		Number int    `json:"importantNumber"`
	}{}
	d, err := jam.NewFileDecoder(ss...)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	if err := d.Decode(&v); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	d.Close()
	fmt.Printf("thing is %s\nother thing is %s\nthird thing is %s\nimportant number is %d\n", v.Thing, v.Other, v.Third, v.Number)
	// Output:
	// thing is moar thing, not default, moar
	// other thing is default other thing
	// third thing is totes extra not even slightly default
	// important number is 57054
}
