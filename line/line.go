package line

import (
	"fmt"
	"io"
	"strings"

	"github.com/stts-se/pronlex/lex"
)

// Field is a simple const for line field definition types
type Field int

//go:generate stringer -type=Field

// TODO add field for Reading(s)?

const (
	// Orth orthography
	Orth Field = iota

	// Pos part-of-speech (noun, verb, NN, VB, etc)
	Pos

	// Morph morphological tags (case, gender, tense, etc)
	Morph

	// WordParts decompounded orthography field (for compounds)
	WordParts

	// Lang the word's language
	Lang

	// Trans1 the primary transcription
	Trans1

	// Translang1 the language of the primary transcription
	Translang1

	// Trans2 transcription variant
	Trans2

	// Translang2 language for Trans2
	Translang2

	// Trans3 transcription variant
	Trans3

	// Translang3 language for Trans3
	Translang3

	// Trans4 transcription variant
	Trans4

	// Translang4 language for Trans4
	Translang4

	// Trans5 transcription variant
	Trans5

	// Translang5 language for Trans5
	Translang5

	// Trans6 transcription variant
	Trans6

	// Translang6 language for Trans6
	Translang6

	// Lemma the lemma form. Ttypically orthographic lemmma + some kind of (disambiguation) identifier, eg., wind_01.
	Lemma

	// Paradigm rule reference (id) for generating inflected forms from lemma
	Paradigm

	// StatusName refers to a status category of the entry, such as 'ok', 'skip' or similar
	StatusName

	// StatusSource refers to the source of a status (user id, reference data id, etc)
	StatusSource

	// Preferred field to use label certain entries preferred over other ones with the same orthography; 1 = preferred, 0 = not preferred; Schema triggers only one preferred per orthographic word
	Preferred

	// Tag is an optional disambiguation tag
	Tag

	// Comments is an optional field
	Comments
)

// FormatTest defines a test to run upon initialization of Format (using NewFormat)
type FormatTest struct {
	InputLine  string
	Fields     map[Field]string
	OutputLine string
}

// Parser is used to define a lexicon's line parser. To implement your own parser, make sure to implement functions Parse(string) and String(map[Field]string)
type Parser interface {

	// Format is the line.Format instance used for line parsing inside of this parser
	Format() Format

	// Parse is used for parsing input lines
	Parse(string) (map[Field]string, error)

	// String is used to generate an output line from a set of fields
	String(map[Field]string) (string, error)

	// Entry2String is used to generate an output line from an input entry
	Entry2String(e lex.Entry) (string, error)
}

// Format is used to define a lexicon's line.
// This a struct for package private usage.
// To create a new Format instance, use NewFormat.
type Format struct {
	Name     string
	FieldSep string
	Fields   map[Field]int
	NFields  int
}

// NewFormat is a public constructor for Format with built-in error checks and tests
func NewFormat(name string, fieldSep string, fields map[Field]int, nFields int, tests []FormatTest) (Format, error) {
	f := Format{name, fieldSep, fields, nFields}
	var errs = make([]string, 0)
	for _, t := range tests {
		fieldsRes, err := f.Parse(t.InputLine)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%v", err))
		} else if !equals(fieldsRes, t.Fields) {
			errs = append(errs, fmt.Sprintf("Format.Parse: expected %v, found %v", t.Fields, fieldsRes))
		}
		lineRes, err := f.String(t.Fields)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%v", err))
		} else if lineRes != t.OutputLine {
			errs = append(errs, fmt.Sprintf("Format.String: expected <%v>, found <%v>", t.OutputLine, lineRes))
		}
	}
	if len(errs) > 0 {
		return Format{}, fmt.Errorf("%s", strings.Join(errs, " : "))
	}
	return f, nil
}

// Parse is used for parsing input lines
func (f Format) Parse(line string) (map[Field]string, error) {
	inputFields := strings.Split(line, f.FieldSep)
	if len(inputFields) != f.NFields {
		return make(map[Field]string), fmt.Errorf("expected %v fields, found %v:\t%s", f.NFields, len(inputFields), line)
	}
	var res = make(map[Field]string)
	for field, i := range f.Fields {
		if i < len(inputFields) {
			res[field] = strings.TrimSpace(inputFields[i])
		} else {
			return make(map[Field]string), fmt.Errorf("couldn't find field %v with index %v in line %v", field, i, line)
		}
	}
	return res, nil
}

func (f Format) Header() string {
	var res = make([]string, f.NFields)
	for field, i := range f.Fields {
		res[i] = field.String()
	}
	return strings.Join(res, f.FieldSep)
}

// String is used to generate an output line from a set of fields
func (f Format) String(fields map[Field]string) (string, error) {
	var res = make([]string, f.NFields)
	for field, s := range fields {
		i, ok := f.Fields[field]
		if ok {
			res[i] = s
		} else {
			return "", fmt.Errorf("undefined field name: %v ", field.String())
		}
	}
	return strings.Join(res, f.FieldSep), nil
}

// FileWriter is used for writing entries to file (using an io.Writer)
type FileWriter struct {
	Parser Parser
	Writer io.Writer
	size   int
}

// Size returns the size of the FileWriter content
func (w FileWriter) Size() int {
	return w.size
}

// Write is used to write one lex.Entry at a time to a file (using an io.Writer)
func (w FileWriter) Write(e lex.Entry) error {
	s, err := w.Parser.Entry2String(e)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w.Writer, "%s\n", s)
	return err
}
