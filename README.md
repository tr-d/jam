# jam

if structured data is scones and you are clotted cream, this is jam


## get

```bash
go get -tags pretty github.com/tr-d/jam/cmd/jam
```


## use
```bash
# usage
jam -h

# detailed help
jam -H

# examples
jam -X
```


### json to yaml
```bash
jam '{"blep":7,"mlem":9}'

# implied
jam -m '{"blep":7,"mlem":9}' -e yaml -o -

# output
blep: 7
mlem: 9
```


### toml to yaml
```bash
jam 'cute = "blep"'

# implied
jam -m 'cute = "blep"' -e yaml -o -

# output
cute: blep
```


### merge
```bash
jam -m '{"blep":2,"mlem":6}' -m '{"blep":4}' -e json

# implied
jam -m '{"blep":2,"mlem":6}' -m '{"blep":4}' -e json -o -

# output
{"blep":4,"mlem":6}
```


### diff
```bash
jam -m '{"blep":2,"mlem":6}' -d '{"blep":4,"mlem":6}' -e json

# implied
jam -m '{"blep":2,"mlem":6}' -d '{"blep":4,"mlem":6}' -e json -o -

# output
{"blep":4}
```


### filter
```bash
jam -m '{"cute":{"blep":3,"mlem":5}}' -f cute.blep

# implied
jam -m '{"cute":{"blep":3,"mlem":5}}' -f cute.blep -e yaml -o -

# output
cute:
  blep: 3
```


### template
```bash
jam -m '["blep","mlem"]' -x '{{range .}}kitty gon {{println .}}{{end}}'

# implied
jam -m '["blep","mlem"]' -x '{{range .}}kitty gon {{println .}}{{end}}' -o -

# output
kitty gon blep
kitty gon mlem
```


### struct
```bash
curl https://api.github.com/user | jam -e struct

# implied
jam -m - -e struct -o -

# output
type T struct {
        DocumentationUrl string `json:"documentation_url"`
        Message          string `json:"message"`
}
```


## package

```go
import "github.com/tr-d/jam"
```


**Decoder** decodes from yaml, json, or toml. The format is detected automatically.

```go
err := jam.NewDecoder(reader).Decode(&v)
```


Go **struct tags** labelled `jam` are evaluated with jmespath.

```go
v := struct {
	X     string   `jam:"foo.x"`
	Y     string   `jam:"foo.y"`
	Names []string `jam:"p[].name"`
}{}

err := jam.NewDecoder(reader).Decode(&v)
```

Jam struct tags work on the decode side only. They are built to play
nice with `json` struct tags. You can use a combination of either or both.
In the case of both, the `jam` struct tag is used by the decoder.

Note: actually, under the hood, the value being decoded is converted to be
json decodable, so both `json` and `jam` tags are in play.


**Encoder** encodes to yaml, json, toml, go syntax or struct.

```go
e := jam.NewEncoder(writer)

// the default is yaml
err := e.Encode(v)

err := e.AsGo().Encode(v)
err := e.AsJson().Encode(v)
err := e.AsToml().Encode(v)
err := e.AsStruct().Encode(v)
err := e.AsYaml().Encode(v)
```


**Core functions**

```go
func Diff(a, b interface{}) interface{}
func Merge(a, b interface{}) interface{}

func Filter(v interface{}, path string) interface{}
func FilterI(v interface{}, path string) interface{}
func FilterR(v interface{}, path string) interface{}
func FilterIR(v interface{}, path string) interface{}
func Query(v interface{}, s string) interface{}
```

