package symbolset

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/stts-se/pronlex/lex"
)

// SymbolSetType is used for accent placement, etc.
type SymbolSetType int

const (
	CMU SymbolSetType = iota
	SAMPA
	IPA
	Other
)

// SymbolCat is used to categorize transcription symbols.
type SymbolCat int

const (
	// Syllabic is used for syllabic phonemes (typically vowels and syllabic consonants)
	Syllabic SymbolCat = iota

	// NonSyllabic is used for non-syllabic phonemes (typically consonants)
	NonSyllabic

	// Stress is used for stress and accent symbols (primary, secondary, tone accents, etc)
	Stress

	// PhonemeDelimiter is used for phoneme delimiters (white space, empty string, etc)
	PhonemeDelimiter

	// SyllableDelimiter is used for syllable delimiters
	SyllableDelimiter

	// MorphemeDelimiter is used for morpheme delimiters that need not align with
	// morpheme boundaries in the decompounded orthography
	MorphemeDelimiter

	// CompoundDelimiter is used for compound delimiters that should be aligned
	// with compound boundaries in the decompounded orthography
	CompoundDelimiter

	// WordDelimiter is used for word delimiters
	WordDelimiter
)

// IPASymbol ipa symbol string with Unicode representation
type IPASymbol struct {
	String  string
	Unicode string
}

// Symbol represent a phoneme, stress or delimiter symbol used in transcriptions, including the IPA symbol with unicode
type Symbol struct {
	String string
	Cat    SymbolCat
	Desc   string
	IPA    IPASymbol
}

// SymbolSet is a struct for package private usage.
// To create a new 'SymbolSet' instance, use NewSymbolSet
type SymbolSet struct {
	Name    string
	Type    SymbolSetType
	Symbols []Symbol

	// to check if the struct has been initialized properly
	isInit bool

	// derived values computed upon initialization
	phonemes        []Symbol
	phoneticSymbols []Symbol
	stressSymbols   []Symbol
	syllabic        []Symbol
	nonSyllabic     []Symbol

	PhonemeRe     *regexp.Regexp
	SyllabicRe    *regexp.Regexp
	NonSyllabicRe *regexp.Regexp
	SymbolRe      *regexp.Regexp

	ipaPhonemeRe     *regexp.Regexp
	ipaSyllabicRe    *regexp.Regexp
	ipaNonSyllabicRe *regexp.Regexp

	phonemeDelimiter          Symbol
	phonemeDelimiterRe        *regexp.Regexp
	repeatedPhonemeDelimiters *regexp.Regexp
}

// ValidSymbol checks if a string is a valid symbol or not
func (ss SymbolSet) ValidSymbol(symbol string) bool {
	return contains(ss.Symbols, symbol)
}

// Get searches the SymbolSet for a symbol with the given string
func (ss SymbolSet) Get(symbol string) (Symbol, error) {
	for _, s := range ss.Symbols {
		if s.String == symbol {
			return s, nil
		}
	}
	return Symbol{}, fmt.Errorf("no symbol /%s/ in symbol set", symbol)
}

// getFromIPA searches the SymbolSet for a symbol with the given IPA symbol string
func (ss SymbolSet) getFromIPA(ipa string) (Symbol, error) {
	for _, s := range ss.Symbols {
		if s.IPA.String == ipa {
			return s, nil
		}
	}
	return Symbol{}, fmt.Errorf("no ipa symbol /%s/ in symbol set", ipa)
}

// SplitTranscription splits the input transcription into separate symbols
func (ss SymbolSet) SplitTranscription(input string) ([]string, error) {
	if !ss.isInit {
		panic("symbolSet " + ss.Name + " has not been initialized properly!")
	}
	delim := ss.phonemeDelimiterRe
	if delim.FindStringIndex("") != nil {
		splitted, unknown, err := splitIntoPhonemes(ss.Symbols, input)
		if err != nil {
			return []string{}, err
		}
		if len(unknown) > 0 {
			return []string{}, fmt.Errorf("found unknown phonemes in transcription /%s/: %v\n", input, unknown)
		}
		return splitted, nil
	}
	return delim.Split(input, -1), nil
}

// SplitIPATranscription splits the input transcription into separate symbols
func (ss SymbolSet) SplitIPATranscription(input string) ([]string, error) {
	if !ss.isInit {
		panic("symbolSet " + ss.Name + " has not been initialized properly!")
	}
	symbols := []Symbol{}
	for _, s := range ss.Symbols {
		ipa := s
		ipa.String = ipa.IPA.String
		symbols = append(symbols, ipa)
	}
	splitted, unknown, err := splitIntoPhonemes(symbols, input)
	if err != nil {
		return []string{}, err
	}
	if len(unknown) > 0 {
		return []string{}, fmt.Errorf("found unknown phonemes in transcription /%s/: %v\n", input, unknown)
	}
	return splitted, nil
}

// ConvertToIPA maps one input transcription string into an IPA transcription
func (ss SymbolSet) ConvertToIPA(trans string) (string, error) {
	res := trans
	res, err := preFilter(ss, trans, ss.Type)
	if err != nil {
		return "", err
	}
	splitted, err := ss.SplitTranscription(res)
	if err != nil {
		return "", err
	}
	var mapped = make([]string, 0)
	for _, fromS := range splitted {
		symbol, err := ss.Get(fromS)
		if err != nil {
			return "", fmt.Errorf("input symbol /%s/ is undefined : %v", fromS, err)
		}
		to := symbol.IPA.String
		if len(to) > 0 {
			mapped = append(mapped, to)
		}
	}
	res = strings.Join(mapped, ss.phonemeDelimiter.IPA.String)

	res, err = postFilter(ss, res, IPA)
	return res, err
}

// ConvertFromIPA maps one input IPA transcription into the current symbol set
func (ss SymbolSet) ConvertFromIPA(trans string) (string, error) {
	res := trans
	res, err := preFilter(ss, trans, IPA)
	if err != nil {
		return "", err
	}
	splitted, err := ss.SplitIPATranscription(res)
	if err != nil {
		return "", err
	}
	var mapped = make([]string, 0)
	for _, fromS := range splitted {
		symbol, err := ss.getFromIPA(fromS)
		if err != nil {
			return "", fmt.Errorf("input symbol /%s/ is undefined : %v", fromS, err)
		}
		to := symbol.String
		if len(to) > 0 {
			mapped = append(mapped, to)
		}
	}
	res = strings.Join(mapped, ss.phonemeDelimiter.String)

	// remove repeated phoneme delimiters, if any
	res = ss.repeatedPhonemeDelimiters.ReplaceAllString(res, ss.phonemeDelimiter.String)
	res, err = postFilter(ss, res, ss.Type)
	return res, err
}

// MapTranscriptions maps the input entry's transcriptions (in-place)
func (ss SymbolSet) MapTranscriptionsToIPA(e *lex.Entry) error {
	var newTs []lex.Transcription
	var errs []string
	for _, t := range e.Transcriptions {
		newT, err := ss.ConvertToIPA(t.Strn)
		if err != nil {
			errs = append(errs, err.Error())
		}
		newTs = append(newTs, lex.Transcription{ID: t.ID, Strn: newT, EntryID: t.EntryID, Language: t.Language, Sources: t.Sources})
	}
	e.Transcriptions = newTs
	if len(errs) > 0 {
		return fmt.Errorf("%v", strings.Join(errs, "; "))
	}
	return nil
}

// MapTranscriptions maps the input entry's transcriptions (in-place)
func (ss SymbolSet) MapTranscriptionsFromIPA(e *lex.Entry) error {
	var newTs []lex.Transcription
	var errs []string
	for _, t := range e.Transcriptions {
		newT, err := ss.ConvertFromIPA(t.Strn)
		if err != nil {
			errs = append(errs, err.Error())
		}
		newTs = append(newTs, lex.Transcription{ID: t.ID, Strn: newT, EntryID: t.EntryID, Language: t.Language, Sources: t.Sources})
	}
	e.Transcriptions = newTs
	if len(errs) > 0 {
		return fmt.Errorf("%v", strings.Join(errs, "; "))
	}
	return nil
}
