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
