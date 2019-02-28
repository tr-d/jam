// +build pretty

package pretty

import (
	"io"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	lexg "github.com/alecthomas/chroma/lexers/g"
	lexj "github.com/alecthomas/chroma/lexers/j"
	"github.com/alecthomas/chroma/styles"
)

// Go makes prettified Go syntax for terminal display
func (b *Buffer) Go() {
	b.appendf(func(w io.Writer, s string) error {
		t, err := lexg.Go.Tokenise(nil, s)
		if err != nil {
			return err
		}
		return formatters.TTY16m.Format(w, goStyle, t)
	})
}

// Yaml makes prettified Yaml for terminal display
func (b *Buffer) Yaml() {
	b.appendf(func(w io.Writer, s string) error {
		t, err := yamlLexer.Tokenise(nil, s)
		if err != nil {
			return err
		}
		return formatters.TTY16m.Format(w, chromaStyle, t)
	})
}

// Json makes prettified Json for terminal display
func (b *Buffer) Json() {
	b.appendf(func(w io.Writer, s string) error {
		t, err := lexj.JSON.Tokenise(nil, s)
		if err != nil {
			return err
		}
		return formatters.TTY16m.Format(w, chromaStyle, t)
	})
}

// Toml makes prettified Toml for terminal display
func (b *Buffer) Toml() {
	b.appendf(func(w io.Writer, s string) error {
		t, err := tomlLexer.Tokenise(nil, s)
		if err != nil {
			return err
		}
		return formatters.TTY16m.Format(w, chromaStyle, t)
	})
}

const (
	pink      = "#ffafc7"
	blue      = "#74d7ec"
	white     = "#fbf9f5"
	lightBlue = "#b7e8f0"
)

var chromaStyle = styles.Register(chroma.MustNewStyle("jam", map[chroma.TokenType]string{
	chroma.Number:      white,
	chroma.Keyword:     white,
	chroma.Date:        white,
	chroma.Comment:     lightBlue,
	chroma.String:      blue,
	chroma.NameTag:     pink,
	chroma.Punctuation: "bold " + white,
	chroma.Error:       pink,
}))

var goStyle = styles.Register(chroma.MustNewStyle("jamgo", map[chroma.TokenType]string{
	chroma.Number:      white,
	chroma.Keyword:     pink,
	chroma.Date:        white,
	chroma.Comment:     lightBlue,
	chroma.String:      blue,
	chroma.NameTag:     pink,
	chroma.Punctuation: "bold " + white,
	chroma.Error:       pink,
}))

var yamlLexer = lexers.Register(chroma.MustNewLexer(
	&chroma.Config{
		Name: "jamyaml",
	},
	chroma.Rules{
		"root": {
			chroma.Include("whitespace"),
			{`^\w+(?=:)`, chroma.NameTag, nil},
			{`:`, chroma.Punctuation, chroma.Push("value")},
			{`-`, chroma.Punctuation, chroma.Push("list")},
		},
		"list": {
			chroma.Include("whitespace"),
			{`^\w+(?=:)`, chroma.NameTag, nil},
			{`:`, chroma.Punctuation, chroma.Push("value")},
			chroma.Include("value"),
		},
		"value": {
			chroma.Include("whitespace"),
			{`([>|])(\s+)((?:(?:.*?$)(?:[\n\r]*?\2)?)*)`, chroma.ByGroups(chroma.StringDoc, chroma.StringDoc, chroma.StringDoc), nil},
			{chroma.Words(``, `\b`, "true", "false", "null"), chroma.KeywordConstant, nil},
			{`\d\d\d\d-\d\d-\d\d([T ]\d\d:\d\d:\d\d(\.\d+)?(Z|\s+[-+]\d+)?)?`, chroma.LiteralDate, nil},
			{`[^\s]*[a-zA-Z\-\.,:;/][^\s]*`, chroma.String, nil},
			{`"(?:\\.|[^"])*"`, chroma.StringDouble, nil},
			{`'(?:\\.|[^'])*'`, chroma.StringSingle, nil},
			{`\b[+\-]?(0x[\da-f]+|0o[0-7]+|(\d+\.?\d*|\.?\d+)(e[\+\-]?\d+)?|\.inf|\.nan)\b`, chroma.Number, nil},
			chroma.Include("punctuation"),
			{`[?,\[\]]`, chroma.Punctuation, nil},
			{`(?=\n)`, chroma.Whitespace, chroma.Pop(1)},
			{`.`, chroma.String, nil},
		},
		"punctuation": {
			{`-`, chroma.Punctuation, nil},
		},
		"whitespace": {
			{` +`, chroma.Whitespace, nil},
			{`#[^\n]+`, chroma.Comment, nil},
		},
	},
))

var tomlLexer = lexers.Register(chroma.MustNewLexer(
	&chroma.Config{
		Name: "jamtoml",
	},
	chroma.Rules{
		"root": {
			chroma.Include("whitespace"),
			{`[\w\.\-]+(?=\]+)`, chroma.NameTag, nil},
			{`[\w\-]+\s*(?==)`, chroma.NameTag, nil},
			chroma.Include("value"),
			{`[=,\[\]]`, chroma.Punctuation, nil},
			{`.`, chroma.Text, nil},
		},
		"value": {
			{chroma.Words(``, `\b`, "true", "false", "null"), chroma.KeywordConstant, nil},
			{`"(?:\\.|[^"])*"`, chroma.StringDouble, nil},
			{`'(?:\\.|[^'])*'`, chroma.StringSingle, nil},
			{`\d\d\d\d-\d\d-\d\d([T ]\d\d:\d\d:\d\d(\.\d+)?(Z|\s+[-+]\d+)?)?`, chroma.LiteralDate, nil},
			{`\b[+\-]?(0x[\da-f]+|0o[0-7]+|(\d+\.?\d*|\.?\d+)(e[\+\-]?\d+)?|\.inf|\.nan)\b`, chroma.Number, nil},
		},
		"whitespace": {
			{`\s+`, chroma.Whitespace, nil},
		},
	},
))
