package lex

import (
	"fmt"
	"io"
	"strings"
)

// DBRef a database reference string (for mariadb: the database name; for sqlite: the database filename without extension)
type DBRef string

// LexName a lexicon name
type LexName string

// LexRef a lexicon reference specified by DBRef and LexName
type LexRef struct {
	DBRef   DBRef   `json:"dbRef,omitempty"`
	LexName LexName `json:"lexName,omitempty"`
}

// LexRefWithInfo is a lexicon reference (LexRef) with additional info (SymbolSetName)
type LexRefWithInfo struct {
	LexRef        LexRef
	SymbolSetName string
}

/*
ParseLexRef is used to parse a lexicon reference string into a LexRef struct

	var fullLexName  = "pronlex:sv-se-nst"
	var lexRef, _    = ParseLexRef(fullLexName)
	// lexRef.DBRef  = pronlex
	// lexRef.LexName = sv-se-nst

*
*/
func ParseLexRef(fullLexName string) (LexRef, error) {
	nameSplit := strings.SplitN(strings.TrimSpace(fullLexName), ":", 2)
	if len(nameSplit) != 2 {
		return LexRef{}, fmt.Errorf("ParseLexRef: failed to split full lexicon name into two colon separated parts: '%s'", fullLexName)
	}
	db := nameSplit[0]
	if db == "" {
		return LexRef{}, fmt.Errorf("ParseLexRef: db part of lexicon name empty: '%s'", fullLexName)
	}
	lex := nameSplit[1]
	if lex == "" {
		return LexRef{}, fmt.Errorf("ParseLexRef: lexicon part of full lexicon name empty: '%s'", fullLexName)
	}

	return NewLexRef(db, lex), nil
}

// NewDBRef creates a database reference from input (downcased) strings
func NewDBRef(dbName string) DBRef {
	return DBRef(strings.ToLower(strings.TrimSpace(dbName)))
}

// NewLexRef creates a lexicon reference from input (downcased) strings
func NewLexRef(lexDB string, lexName string) LexRef {
	return LexRef{DBRef: DBRef(strings.ToLower(strings.TrimSpace(lexDB))),
		LexName: LexName(strings.ToLower(strings.TrimSpace(lexName))),
	}
}

// NewLexRefWithInfo creates a lexicon reference with symbol set, from (downcased) input strings
func NewLexRefWithInfo(lexDB string, lexName string, symbolSetName string) LexRefWithInfo {
	lexRef := NewLexRef(lexDB, lexName)
	return LexRefWithInfo{LexRef: lexRef, SymbolSetName: strings.ToLower(strings.TrimSpace(symbolSetName))}
}

func (lr LexRef) String() string {
	return fmt.Sprintf("%s:%s", string(lr.DBRef), string(lr.LexName))
}

// EntryStatus associates a status to an Entry. The status has a name (such as 'ok') and a source (a string identifying who or what generated the status)
type EntryStatus struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Source string `json:"source,omitempty"`
	//EntryID int64  `json:"entryId"`
	//Timestamp int64  `json:"timestamp"`
	Timestamp string `json:"timestamp,omitempty"`
	Current   bool   `json:"current,omitempty"`
}

type EntryComment struct {
	ID      int64  `json:"id,omitempty"`
	EntryID int64  `json:"entryId,omitempty"`
	Source  string `json:"source,omitempty"`
	Label   string `json:"label,omitempty"`
	Comment string `json:"comment,omitempty"`
}

func (c EntryComment) String() string {
	return fmt.Sprintf("[%s: %s] (%s)", c.Label, c.Comment, c.Source)
}

// EntryValidation associates a validation result to an Entry
type EntryValidation struct {
	ID int64 `json:"id,omitempty"`

	// Lower case name of level of severity
	Level     string `json:"level"`
	RuleName  string `json:"ruleName"`
	Message   string `json:"Message"`
	Timestamp string `json:"timestamp"`
}

func (ev EntryValidation) String() string {
	if ev.Timestamp != "" {
		return fmt.Sprintf("%s|%s: %s (%v)", ev.Level, ev.RuleName, ev.Message, ev.Timestamp)
	}
	return fmt.Sprintf("%s|%s: %s", ev.Level, ev.RuleName, ev.Message)
}

// SourceDelimiter is used to split a string of sevaral sources into a slice
var SourceDelimiter = " : "

// Transcription corresponds to the transcription db table
type Transcription struct {
	ID       int64    `json:"id,omitempty"`
	EntryID  int64    `json:"entryId,omitempty"`
	Strn     string   `json:"strn"`
	Language string   `json:"language,omitempty"`
	Sources  []string `json:"sources,omitempty"`
}

// AddSource ... adds a source string at the beginning of the
// Transcription.Sources slice. If the source is already present,
// AddSource silently ignores to add the already existing
// source. AddSource returns an error when the input string contains the
// SourceDelimiter string.
func (t *Transcription) AddSource(s string) error {
	sDC := strings.ToLower(strings.TrimSpace(s))
	if strings.Contains(sDC, SourceDelimiter) {
		return fmt.Errorf("cannot add source containing the source delimiter : '%v'", SourceDelimiter)
	}

	for i := range t.Sources {
		if sDC == t.Sources[i] {
			return nil // source already there
		}
	}

	t.Sources = append([]string{sDC}, t.Sources...)

	return nil
}

// SourcesString returns the []string items of Transcription.Sources as a string, where the items are delimited by SourceDelimiter
func (t Transcription) SourcesString() string {
	return strings.Join(t.Sources, SourceDelimiter)
}

// TranscriptionSlice is used for
// soring according to ascending id
type TranscriptionSlice []Transcription

func (a TranscriptionSlice) Len() int           { return len(a) }
func (a TranscriptionSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TranscriptionSlice) Less(i, j int) bool { return a[i].ID < a[j].ID }

// Lemma corresponds to a row of the lemma db table
type Lemma struct {
	ID       int64  `json:"id,omitempty"`
	Strn     string `json:"strn,omitempty"`
	Reading  string `json:"reading,omitempty"`
	Paradigm string `json:"paradigm,omitempty"`
}

// Entry defines a lexical entry. It does not correspond one-to-one to
// the entry db table, since it contains data also from associated
// tables (Lemma, Tag, Transcription, EntryValidations).  The Tag
// field holds an arbitrary, optional, lower case string to
// disambiguate between different lex.Entries charing the same
// othograpy. Two different lex.Entries cannot have identical
// lex.Entry.Tags (the database should not allow this).
type Entry struct {
	ID               int64             `json:"id,omitempty"`
	LexRef           LexRef            `json:"lexRef,omitempty"`
	Strn             string            `json:"strn"`
	Language         string            `json:"language,omitempty"`
	PartOfSpeech     string            `json:"partOfSpeech,omitempty"`
	Morphology       string            `json:"morphology,omitempty"`
	WordParts        string            `json:"wordParts,omitempty"`
	Lemma            Lemma             `json:"lemma,omitempty"`
	Transcriptions   []Transcription   `json:"transcriptions"`
	EntryStatus      EntryStatus       `json:"status,omitempty"` // TODO Probably should be a slice of statuses?
	EntryValidations []EntryValidation `json:"entryValidations,omitempty"`

	// Preferred flag: 1=true, 0=false; schema triggers only one preferred per orthographic word
	//Preferred        int64             `json:"preferred"`
	Preferred bool           `json:"preferred,omitempty"`
	Tag       string         `json:"tag,omitempty"`
	Comments  []EntryComment `json:"comments,omitempty"`
}

// EntryWriter is an interface defining things to which one can write an Entry.
// See EntrySliceWriter, for returning a slice of Entry, and EntryFileWriter, for writing Entries to file.
type EntryWriter interface {
	Write(Entry) error
	Size() int
}

// EntryFileWriter outputs formated entries to an io.Writer.
// Example usage:
//
//	bf := bufio.NewWriter(f)
//	defer bf.Flush()
//	bfx := lex.EntriesFileWriter{bf}
//	dbapi.LookUp(db, q, bfx)
type EntryFileWriter struct {
	size   int
	Writer io.Writer
}

// Size returns the size of the EntryFileWriter content
func (w *EntryFileWriter) Size() int {
	return w.size
}

// Write is used to write one lex.Entry at a time to a file
func (w *EntryFileWriter) Write(e Entry) error {
	// TODO call to line formatting of Entry
	w.size = w.size + 1
	_, err := fmt.Fprintf(w.Writer, "%v\n", e)
	return err
}

// EntrySliceWriter is a container for returning Entries from a LookUp call to the db
// Example usage:
//
//	var q := dbapi.Query{ ... }
//	var esw lex.EntrySliceWriter
//	err := dbapi.LookUp(db, q, &esw)
//	[...] esw.Entries // process Entries
type EntrySliceWriter struct {
	Entries []Entry
}

// Size returns the size of the EntryFileWriter content
func (w *EntrySliceWriter) Size() int {
	return len(w.Entries)
}

// Write is used to write one lex.Entry at a time to a file
func (w *EntrySliceWriter) Write(e Entry) error {
	w.Entries = append(w.Entries, e)
	return nil
}
