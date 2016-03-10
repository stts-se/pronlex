// Package symbolset is used to define symbol sets and mapping between different sets, such as IPA to SAMPA, and so on.
package symbolset

import (
	"fmt"
	"regexp"
	"strings"
)

// structs in package symbolset

// SymbolSet is a struct for package private usage.
// To create a new SymbolSet, use NewSymbolSet
type SymbolSet struct {
	Name    string
	Symbols []Symbol

	// derived values computed upon initialization
	phonemes                  []Symbol
	phoneticSymbols           []Symbol
	stressSymbols             []Symbol
	syllabic                  []Symbol
	nonSyllabic               []Symbol
	phonemeDelimiters         []Symbol
	explicitPhonemeDelimiters []Symbol

	phonemeDelimiter            Symbol
	explicitPhonemeDelimiter    Symbol
	hasExplicitPhonemeDelimiter bool

	phonemeRe          *regexp.Regexp
	syllabicRe         *regexp.Regexp
	nonSyllabicRe      *regexp.Regexp
	phonemeDelimiterRe *regexp.Regexp
	symbolRe           *regexp.Regexp
}

// Mapper is a struct for package private usage.
// To create a new Mapper, use NewMapper.
type Mapper struct {
	FromName string
	ToName   string
	Symbols  []SymbolPair

	ipa ipa

	// derived values computed upon initialization
	fromIsIPA bool
	toIsIPA   bool

	from      SymbolSet
	to        SymbolSet
	symbolMap map[Symbol]Symbol

	repeatedPhonemeDelimiters *regexp.Regexp
}

func (m Mapper) preFilter(trans string, ss SymbolSet) (string, error) {
	switch m.fromIsIPA {
	case true:
		return m.ipa.filterBeforeMappingFromIpa(trans, ss)
	default:
		return trans, nil
	}
}

func (m Mapper) postFilter(trans string, ss SymbolSet) (string, error) {
	switch m.toIsIPA {
	case true:
		return m.ipa.filterAfterMappingToIpa(trans, ss)
	default:
		return trans, nil
	}
}

func (m Mapper) mapTranscription(input string) (string, error) {
	res, err := m.preFilter(input, m.from)
	if err != nil {
		return "", err
	}
	splitted, err := m.from.SplitTranscription(res)
	if err != nil {
		return "", err
	}
	var mapped = make([]string, 0)
	for _, fromS := range splitted {
		from, err := m.from.Get(fromS)
		if err != nil {
			return "", fmt.Errorf("input symbol /%s/ is undefined : %v", fromS, err)
		}
		to := m.symbolMap[from]
		//if to.Type == UndefinedSymbol {
		//	return "", fmt.Errorf("couldn't map symbol /%s/", fromS)
		//}
		if len(to.String) > 0 {
			mapped = append(mapped, to.String)
		}
	}
	mapped, err = m.to.filterAmbiguous(mapped)
	if err != nil {
		return "", err
	}
	res = strings.Join(mapped, m.to.phonemeDelimiter.String)

	// remove repeated phoneme delimiters
	res = m.repeatedPhonemeDelimiters.ReplaceAllString(res, m.to.phonemeDelimiter.String)
	return m.postFilter(res, m.to)
}

// SymbolPair is a tuple inside the Mapper used to store the symbol mappings in a list in preserved order
type SymbolPair struct {
	Sym1 Symbol
	Sym2 Symbol
}

// SymbolSlice is used for sorting slices of symbols according to symbol length. Symbols with equal length will be sorted alphabetically.
type symbolSlice []Symbol

func (a symbolSlice) Len() int      { return len(a) }
func (a symbolSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a symbolSlice) Less(i, j int) bool {
	if len(a[i].String) != len(a[j].String) {
		return len(a[i].String) > len(a[j].String)
	}
	return a[i].String < a[j].String
}

// SymbolType is used to categorize transcription symbols.
type SymbolType int

const (
	// Syllabic is used for syllabic phonemes (typically vowels and syllabic consonants)
	Syllabic SymbolType = iota

	// NonSyllabic is used for non-syllabic phonemes (typically consonants)
	NonSyllabic

	// Stress is used for stress and accent symbols (primary, secondary, tone accents, etc)
	Stress

	// PhonemeDelimiter is used for phoneme delimiters such (white space, empty string, etc)
	PhonemeDelimiter

	// ExplicitPhonemeDelimiter is used for explicit phoneme delimiters when
	// the default PhonemeDelimiter is empty, and there are ambiguous phonemes
	// in the symbol set (for example Swedish SAMPA rt vs. r-t)
	ExplicitPhonemeDelimiter

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

// Symbol represent a phoneme, stress or delimiter symbol used in transcriptions
type Symbol struct {
	String string
	Type   SymbolType
	Desc   string
}

// Contains checks if the SymbolSet contains a certain symbol string
func (ss SymbolSet) Contains(symbol string) bool {
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

func (ss SymbolSet) filterAmbiguous(trans []string) ([]string, error) {
	potentiallyAmbs := ss.phoneticSymbols
	phnDel := ss.phonemeDelimiter.String
	explicitPhnDel := ss.explicitPhonemeDelimiter
	var res = make([]string, 0)
	for i, phn0 := range trans[0 : len(trans)-1] {
		phn1 := trans[i+1]
		test := phn0 + phnDel + phn1
		if contains(potentiallyAmbs, test) {
			if !ss.hasExplicitPhonemeDelimiter {
				return nil, fmt.Errorf("explicit phoneme delimiter in %s was undefined when needed for [%s] vs [%s] + [%s]", ss.Name, (phn0 + phn1), phn0, phn1)
			}
			res = append(res, phn0+explicitPhnDel.String)
		} else {
			res = append(res, phn0)
		}
	}
	res = append(res, trans[len(trans)-1])
	return res, nil
}

func (ss SymbolSet) preCheckAmbiguous() error {
	allSymbols := ss.phonemes
	var res = make([]string, 0)
	for _, a := range allSymbols {
		for _, b := range allSymbols {
			if len(a.String) > 0 && len(b.String) > 0 {
				res = append(res, a.String)
				res = append(res, b.String)
			}
		}
	}
	_, err := ss.filterAmbiguous(res)
	return err
}

// SplitTranscription splits the input transcription into separate symbols
func (ss SymbolSet) SplitTranscription(input string) ([]string, error) {
	delim := ss.phonemeDelimiterRe
	if delim.FindStringIndex("") != nil {
		rest := input
		var acc = make([]string, 0)
		for len(rest) > 0 {
			match := ss.symbolRe.FindStringIndex(rest)
			switch match {
			case nil:
				return nil, fmt.Errorf("transcription not splittable (invalid symbols?)! input=/%s/, acc=/%s/, rest=/%s/", input, strings.Join(acc, ss.phonemeDelimiter.String), rest)

			default:
				if match[0] != 0 {
					return nil, fmt.Errorf("couldn't parse transcription /%s/, it may contain undefined symbols", input)
				}
				acc = append(acc, rest[match[0]:match[1]])
				rest = rest[match[1]:]
			}
		}
		return acc, nil
	}
	return delim.Split(input, -1), nil
}

// ipa utilility functions with struct for package private usage.
// To create a new SymbolSet, use NewSymbolSet.
// Symbols and codes: http://www.phon.ucl.ac.uk/home/wells/ipa-unicode.htm#numbers
type ipa struct {
	ipa      string
	accentI  string
	accentII string
}

func (ipa ipa) isIPA(symbolSetName string) bool {
	return strings.Contains(strings.ToLower(symbolSetName), ipa.ipa)
}

func (ipa ipa) checkFilterRequirements(ss SymbolSet) error {
	if !ss.Contains(ipa.accentI) {
		return fmt.Errorf("no IPA stress symbol in stress symbol list? IPA stress =/%v/, stress symbols=%v", ipa.accentI, ss.stressSymbols)
	} else if !ss.Contains(ipa.accentI + ipa.accentII) {
		return fmt.Errorf("no IPA tone II symbol in stress symbol list? IPA stress =/%s/, stress symbols=%v", ipa.accentI, ss.stressSymbols)
	} else {
		return nil
	}
}

func (ipa ipa) filterBeforeMappingFromIpa(trans string, ss SymbolSet) (string, error) {
	// IPA: ˈba`ŋ.ka => ˈ`baŋ.ka"
	err := ipa.checkFilterRequirements(ss)
	if err != nil {
		return "", err
	}
	s := ipa.accentI + "(" + ss.phonemeRe.String() + "+)" + ipa.accentII
	repl, err := regexp.Compile(s)
	if err != nil {
		err = fmt.Errorf("couldn't compile regexp from string '%s' : %v", s, err)
	}
	res := repl.ReplaceAllString(trans, ipa.accentI+ipa.accentII+"$1")
	return res, nil
}

func (ipa ipa) filterAfterMappingToIpa(trans string, ss SymbolSet) (string, error) {
	conditionForAfterMapping := ipa.accentI + ipa.accentII
	// IPA: /'`pa.pa/ => /'pa`.pa/
	if strings.Contains(trans, conditionForAfterMapping) {
		err := ipa.checkFilterRequirements(ss)
		if err != nil {
			return "", err
		}
		repl, err := regexp.Compile(ipa.accentI + ipa.accentII + "(" + ss.nonSyllabicRe.String() + "*)(" + ss.syllabicRe.String() + ")")
		res := repl.ReplaceAllString(trans, ipa.accentI+"$1$2"+ipa.accentII)
		return res, nil
	}
	return trans, nil
}
