package main // import "github.com/tr-d/jam/cmd/jam"

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/tr-d/jam"
	"github.com/tr-d/jam/pretty"
)

// everything is an op
type op struct {
	p  string
	fn *func(*jam.Jam, *pretty.Buffer, string) error
}

// opsvalue is a flag.Value that appends op to slice of ops.
type opsvalue struct {
	ops *[]op
	fn  *func(*jam.Jam, *pretty.Buffer, string) error
}

func (v *opsvalue) Set(s string) error {
	*v.ops = append(*v.ops, op{p: s, fn: v.fn})
	return nil
}

func (v *opsvalue) String() string {
	if v.ops == nil {
		return "[]"
	}
	return fmt.Sprintf("%+v", *v.ops)
}

var (
	opout = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		if b.Len() == 0 {
			if err := openc(j, b, ""); err != nil {
				return err
			}
		}

		var w *os.File
		switch p {
		case "-":
			w = os.Stdout
		default:
			f, err := os.Create(p)
			if err != nil {
				return err
			}
			defer f.Close()
			w = f
		}
		if terminal.IsTerminal(int(w.Fd())) {
			return b.Format(w)
		}
		_, err := b.WriteTo(w)
		return err
	}

	openc = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		e := jam.NewEncoder(b)
		switch {
		case p == "j" || p == "json":
			b.Json()
			e = e.AsJson()
		case p == "t" || p == "toml":
			if len(j.Values()) > 1 {
				return errors.New("multiple objects were decoded from json or yaml, encoding as toml is not supported")
			}
			b.Toml()
			e = e.AsToml()
		case p == "g" || p == "go":
			b.Go()
			e = e.AsGo()
		case p == "s" || p == "struct":
			b.Go()
			e = e.AsStruct()
		default:
			b.Yaml()
		}
		for _, v := range j.Values() {
			if err := e.Encode(v); err != nil {
				return err
			}
		}
		return nil
	}

	opexec = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		switch {
		case len(p) == 0 || p[0] == '\'' || p[0] == '"':
			b.Ugly()
		case strings.Contains(p, ".go.") || strings.HasSuffix(p, ".go"):
			b.Go()
		case strings.Contains(p, ".yml.") || strings.HasSuffix(p, ".yml"):
			b.Yaml()
		case strings.Contains(p, ".json.") || strings.HasSuffix(p, ".json"):
			b.Json()
		case strings.Contains(p, ".toml.") || strings.HasSuffix(p, ".toml"):
			b.Toml()
		default:
			b.Ugly()
		}
		rc, err := source(p)
		if err != nil {
			return fmt.Errorf("exec: %s", err)
		}
		if err := j.Exec(b, rc); err != nil {
			return fmt.Errorf("exec: %s", err)
		}
		return nil
	}

	opmerg = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		vs, err := decode(source(p))
		if err != nil {
			return fmt.Errorf("merge: %s", err)
		}
		j.Merge(vs...)
		return nil
	}

	opdiff = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		vs, err := decode(source(p))
		if err != nil {
			return fmt.Errorf("diff: %s", err)
		}
		j.Diff(vs...)
		return nil
	}

	opflt = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		j.Filter(p)
		return nil
	}

	opflti = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		j.FilterI(p)
		return nil
	}

	opfltr = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		j.FilterR(p)
		return nil
	}

	opfltir = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		j.FilterIR(p)
		return nil
	}

	opqry = func(j *jam.Jam, b *pretty.Buffer, p string) error {
		j.Query(p)
		return nil
	}
)

func source(s string) (io.ReadCloser, error) {
	switch {
	case s == "":
	case s == "-":
		return ioutil.NopCloser(os.Stdin), nil
	case s[0] == '@':
		f, err := os.Open(s[1:])
		if err != nil {
			return nil, fmt.Errorf("%s: %s", s, err)
		}
		return f, nil
	}
	return ioutil.NopCloser(strings.NewReader(s)), nil
}

func decode(rc io.ReadCloser, err error) ([]interface{}, error) {
	vs := []interface{}{}
	if err != nil {
		return vs, err
	}
	defer rc.Close()
	d := jam.NewDecoder(rc)
	for {
		var v interface{}
		err := d.Decode(&v)
		if jam.IsNoMore(err) {
			break
		}
		if err != nil {
			return vs, err
		}
		vs = append(vs, v)
	}
	return vs, nil
}

var opflags = []struct {
	name  string
	fn    *func(*jam.Jam, *pretty.Buffer, string) error
	usage string
}{
	{"d", &opdiff, "diff `in`put (-, @file, string) (yaml, json, toml)"},
	{"m", &opmerg, "merge `in`put (-, @file, string) (yaml, json, toml)"},
	{"x", &opexec, "exec template `in`put to buffer (-, @file, string) (text/template)"},
	{},
	{"e", &openc, "`enc`ode to buffer (yaml, json, toml, go, struct)"},
	{"o", &opout, "write `out` buffer (-, file)"},
	{},
	{"f", &opflt, "`filt`er plain"},
	{"F", &opflti, "`filt`er inverted"},
	{"q", &opqry, "jmespath `query`"},
	{"r", &opfltr, "`filt`er recursive"},
	{"R", &opfltir, "`filt`er recursive inverted"},
}

func main() {
	var (
		h, x, v bool
		ops     = []op{}
	)
	for _, o := range opflags {
		if o.fn != nil {
			flag.Var(&opsvalue{&ops, o.fn}, o.name, o.usage)
		}
	}
	flag.BoolVar(&h, "H", false, "")
	flag.BoolVar(&x, "X", false, "")
	flag.BoolVar(&v, "v", false, "")
	flag.Usage = usage
	flag.Parse()

	switch {
	case h:
		halpsPrint(halpsMoar)
		os.Exit(2)
	case x:
		halpsPrint(exemples)
		os.Exit(2)
	case v:
		versionPrint()
		return
	}

	pops := make([]op, flag.NArg())
	for i, arg := range flag.Args() {
		pops[i] = op{p: arg, fn: &opmerg}
	}
	ops = append(pops, ops...)

	i, o := true, true
	for _, op := range ops {
		switch {
		case op.fn == &opout:
			o = false
		case op.fn == &opmerg || op.fn == &opdiff:
			i, o = false, true
		default:
			o = true
		}
	}
	if i {
		ops = append([]op{{p: "-", fn: &opmerg}}, ops...)
	}
	if o {
		ops = append(ops, op{p: "-", fn: &opout})
	}

	log.SetFlags(0)
	var (
		pb pretty.Buffer
		j  = &jam.Jam{}
	)
	for _, o := range ops {
		f := *o.fn
		if err := f(j, &pb, o.p); err != nil {
			log.Fatal("Error: ", err)
		}
	}
}

var (
	arg0    = filepath.Base(os.Args[0])
	version = "unknown"
)

const (
	cmdFlags = `
  -h	halps
  -H	moar halps
  -X	les exemples
  -v	version

`

	halps = `Usage of %[1]s:

    	%[1]s [flags]

Flags:
`

	halpsMoar = `
Pipeline:
  Flags form a processing pipeline for a generic data tree and an output
  buffer.  Flags may modify the tree or write to the buffer.  Flags are
  applied from left to right.

Inputs (in):
  Inputs are @file, - for stdin, or a literal string.

  Merge (-m <in>) takes one input and merges it with the tree. The input
  will overwrite matching parts of the tree.  Input format may be yaml,
  json or toml.  The format will be detected automatically.

  Diff (-d <in>) is the transpose of merge.  Only the parts of the input that
  are not in the tree will remain.  Input format may be yaml, json or toml.
  The format will be detected automatically.

  Exec (-x <in>) executes a go text template input against the tree.  Input
  format must be a valid go text template.  See https://godoc.org/text/template

Encoding (enc):
  Encoding (-e <enc>) writes yaml, json, toml, go, or struct to the output
  buffer.  Values y, j, t, g, s, are also acceptable if you are feeling lazy.

Outputs (out):
  Output (-o <out>) goes to file or stdout (-). If nothing has been written
  to the ouput buffer, an implicit encode to yaml occurs (-e "yaml").

Queries:
  Query (-q <query>) applies a JMESPath query to the tree. See
  http://jmespath.org/

Filters (filt):
  Queries extract from or otherwise transform the tree.  In contrast, filters
  discard parts of the tree according to the filter query.  They use a path
  syntax which will look familiar and is absolutely simple.

  Map keys are addressed by name.  Nested keys are separated by a dot ".",
  a star "*" matches any key.

  	MapKey "." NestedMapKey
  	"*" "." MapKey

  Lists are addressed with standard index notation.  Lists can be sliced;
  slice notation plays by Go rules.

  	"[]"
  	"[" Index "]"
  	"[" GEIndex? ":" LTIndex? "]"

  A filter query can match a specific value, which follows a "==" and may
  only appear once at the end of the query string.  The value is decoded
  from yaml, json or toml.

  	"==" Value(json, yaml, toml)

  There are four filter behaviours.

  * Plain (-f <filt>)
    + The query matches from the root of the tree.
    + Everything not matching is discarded.

  * Inverted (-F <filt>)
    + The query matches from the root of the tree.
    + Everything matching is discarded.

  * Recursive (-r <filt>)
    + The query matches from anywhere in the tree.
    + Everything not matching is discarded.

  * Recursive Inverted (-R <filt>)
    + The query matches from anywhere in the tree.
    + Everything matching is discarded.

Implicit Flags:
  In certain cases, %[1]s will make some assumtions about what you want to do.

  Pipelines must start with a merge (-m <in>) or diff (-d <in>). Both merge
  and diff on a nil tree produce the same result, the tree is set to the input.
  If the pipeline does not start with one of these, merge from standard input
  (-m "-") is prepended to the flags.

  If the pipeline does not end with an output, output to standard output
  (-o "-") is appended to the flags.

  If the pipeline wants to output (-o <out>) something, but nothing could
  have been written to the output buffer by the previous flags, encode as yaml
  (-e "yaml") is inserted before the output flag.

  Putting all that together, invoking %[1]s with no flags or arguments is
  equivalent to

  	%[1]s -m - -e yaml -o -

  or, merge standard input, encode yaml, write standard output.

  Positional arguments passed after the flags are prepended to the flags as
  merge (-m <in>).

  	%[1]s -q blep @file1 @file2

  is equivalent to

  	%[1]s -m @file1 -m @file2 -q blep

  or, merge file1, merge file2, query blep.
`

	exemples = `
  Les exemples sont très nécessaires!

Convert json to yaml:
  %[1]s '{"blep":7,"mlem":9}'

  # implied: %[1]s -m '{"blep":7,"mlem":9}' -e yaml -o -
  # output:
  # blep: 7
  # mlem: 9

Convert toml to yaml:
  %[1]s 'cute = "blep"'

  # implied: %[1]s -m 'cute = "blep"' -e yaml -o -
  # output:
  # cute: blep

Merge:
  %[1]s -m '{"blep":2,"mlem":6}' -m '{"blep":4}' -e json

  # implied: %[1]s -m '{"blep":2,"mlem":6}' -m '{"blep":4}' -e json -o -
  # output:
  # {"blep":4,"mlem":6}

Diff:
  %[1]s -m '{"blep":2,"mlem":6}' -d '{"blep":4,"mlem":6}' -e json

  # implied: %[1]s -m '{"blep":2,"mlem":6}' -d '{"blep":4,"mlem":6}' -e json -o -
  # output:
  # {"blep":4}

Filter:
  %[1]s -m '{"cute":{"blep":3,"mlem":5}}' -f cute.blep

  # implied: %[1]s -m '{"cute":{"blep":3,"mlem":5}}' -f cute.blep -e yaml -o -
  # output:
  # cute:
  #   blep: 3

Template:
  %[1]s -m '["blep","mlem"]' -x '{{range .}}kitty gon {{println .}}{{end}}'

  # implied:
  # %[1]s -m '["blep","mlem"]' -x '{{range .}}kitty gon {{println .}}{{end}}' -o -
  # output:
  # kitty gon blep
  # kitty gon mlem

Convert to struct:
  curl https://api.github.com/user | %[1]s -e struct

  # implied: %[1]s -m - -e struct -o -
  # output:
  # type T struct {
  #         DocumentationUrl string ` + "`" + `json:"documentation_url"` + "`" + `
  #         Message          string ` + "`" + `json:"message"` + "`" + `
  # }
`
)

func usage() {
	fo := flag.CommandLine.Output()
	fmt.Fprintf(fo, halps, arg0)
	for _, o := range opflags {
		if o.fn != nil {
			t, u := flag.UnquoteUsage(&flag.Flag{Usage: o.usage})
			fmt.Fprintf(fo, "  -%s <%s>\n    \t%s", o.name, t, u)
		}
		fmt.Fprintln(fo)
	}
	fmt.Fprint(fo, cmdFlags)
}
