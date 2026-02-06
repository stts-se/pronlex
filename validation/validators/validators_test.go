package validators

import (
	"testing"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/symbolset"
)

var fsExp = "Expected: '%v' got: '%v'"

// SvNSTHardWired is a temporary function that should not be used in production
func newNSTSvHardWired_ForTesting() (symbolset.SymbolSet, error) {
	name := "sv.se.nst-SAMPA"

	syms := []symbolset.Symbol{
		{Desc: "sil", String: "i:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "sill", String: "I", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "full", String: "u0", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "ful", String: "}:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "matt", String: "a", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "mat", String: "A:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "bot", String: "u:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "bott", String: "U", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "häl", String: "E:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "häll", String: "E", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "aula", String: "au", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "syl", String: "y:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "syll", String: "Y", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "hel", String: "e:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "herr,hett", String: "e", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "nöt", String: "2:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "mött,förra", String: "9", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "mål", String: "o:", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "moll,håll", String: "O", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "bättre", String: "@", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "europa", String: "eu", Cat: symbolset.Syllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "pol", String: "p", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "bok", String: "b", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "tok", String: "t", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "bort", String: "rt", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "mod", String: "m", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "nod", String: "n", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "dop", String: "d", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "bord", String: "rd", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "fot", String: "k", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "våt", String: "g", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "lång", String: "N", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "forna", String: "rn", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "fot", String: "f", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "våt", String: "v", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "kjol", String: "C", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "fors", String: "rs", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "rov", String: "r", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "lov", String: "l", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "sot", String: "s", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "sjok", String: "x", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "hot", String: "h", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "porla", String: "rl", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "jord", String: "j", Cat: symbolset.NonSyllabic, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "syllable delimiter", String: ".", Cat: symbolset.SyllableDelimiter, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "accent I", String: `"`, Cat: symbolset.Stress, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "accent II", String: `""`, Cat: symbolset.Stress, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "secondary stress", String: "%", Cat: symbolset.Stress, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{Desc: "phoneme delimiter", String: " ", Cat: symbolset.PhonemeDelimiter, IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "+", Cat: symbolset.CompoundDelimiter, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
	}

	return symbolset.NewSymbolSet(name, syms)

}

/*
func newNSTNbvHardWired_ForTesting() (symbolset.SymbolSet, error) {
	name := "NST nob sampa"
	symbols := []symbolset.Symbol{
		{String: "@", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "A", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "E", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "I", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "O", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "U", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "u0", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "Y", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\\{", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "9", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "A:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "e:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "i:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "o:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "u:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\\}:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "y:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "{:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "2:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "9:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\\{\\*I", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "9\\*Y", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "A\\*I", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "E\\*\\}", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "O\\*Y", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "o~", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "n=", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "l=", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "n`=", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "l`=", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "}*I", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "a\\*U", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "@\\*U", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "e~", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "3:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "a", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "a:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "U:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "V", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "U4", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "I@", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},

		{String: "p", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "t", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "k", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "b", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "d", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "g", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "f", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "v", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "h", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "j", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "s", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "l", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "r", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "n", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "m", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "N", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "t`", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "d`", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "s`", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "n`", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "l`", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "S", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "C", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "tS", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "dZ", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "w", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "x", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "T", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "D", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "r3", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "Z", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "z", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},

		{String: "%", Cat: symbolset.Stress, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\"\"", Cat: symbolset.Stress, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\"", Cat: symbolset.Stress, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: " ", Cat: symbolset.PhonemeDelimiter, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "$", Cat: symbolset.SyllableDelimiter, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "-", Cat: symbolset.CompoundDelimiter, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
	}
	return symbolset.NewSymbolSet(name, symbols)
}
*/

func TestWhitespace(t *testing.T) {
	symbolset, err := newNSTSvHardWired_ForTesting()
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	vali, err := newSvSeNstValidator(symbolset)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	testRes, err := vali.RunTests()
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if len(testRes.AcceptErrors) > 0 {
		t.Errorf("%v", testRes.AcceptErrors)
	}
	if len(testRes.RejectErrors) > 0 {
		t.Errorf("%v", testRes.RejectErrors)
	}
	if len(testRes.CrossErrors) > 0 {
		t.Errorf("%v", testRes.CrossErrors)
	}
	// fmt.Println(testRes.AcceptErrors)
	// fmt.Println(testRes.RejectErrors)
	// fmt.Println(testRes.CrossErrors)

	var e = lex.Entry{
		Strn:         "banen",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "banen",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\" b A: . n @ n",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	var result = e.EntryValidations

	var expect = []lex.EntryValidation{}
	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "banen",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "banen",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\" b A: . n @ Ö ",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{
		{
			RuleName: "syllabic",
			Level:    "Format",
			Message:  "[...]"},
		{
			RuleName: "SymbolSet",
			Level:    "fatal",
			Message:  "[...]"},
	}
	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "banen",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "banen",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\" b A: . n @  n",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{
		{
			RuleName: "SymbolSet",
			Level:    "Fatal",
			Message:  "[...]"},
	}
	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//

}

func TestRepeated(t *testing.T) {
	symbolset, err := newNSTSvHardWired_ForTesting()
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	vali, err := newSvSeNstValidator(symbolset)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	var e = lex.Entry{
		Strn:         "banen",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "banen",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\" b A: . n @ n",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	var result = e.EntryValidations

	var expect = []lex.EntryValidation{}
	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "banen",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "banen",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\" b A: n . n @ n",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{
		{
			RuleName: "repeated_phonemes",
			Level:    "Fatal",
			Message:  "[...]"},
	}
	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//
}

func TestNst1(t *testing.T) {
	symbolset, err := newNSTSvHardWired_ForTesting()
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	vali, err := newSvSeNstValidator(symbolset)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	var e = lex.Entry{
		Strn:         "banen",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "banen",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\" b A: . n @ n",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	var result = e.EntryValidations

	var expect = []lex.EntryValidation{}
	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "bantorget",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "ban+torget",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\"\" b A: n + % t O r . j @ t",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{}

	if len(result) != len(expect) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "battorget",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "bat+torget",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\"\" b A: t + % t O r . j @ t",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{}

	if len(result) != len(expect) || (len(expect) > 0 && result[0].RuleName != expect[0].RuleName) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "battorget",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "batt+torget",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\"\" b a t + % t O r . j @ t",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{}

	if len(result) != len(expect) || (len(expect) > 0 && result[0].RuleName != expect[0].RuleName) {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "batttorget",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "batt+torget",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "\"\" b a t + % t O r . j @ t",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{
		{
			RuleName: "Decomp2Orth",
			Level:    "Fatal",
			Message:  "[...]"},
	}

	if len(result) != len(expect) || result[0].RuleName != expect[0].RuleName {
		t.Errorf(fsExp, expect, result)
	}

	//

	e = lex.Entry{
		Strn:         "apnos",
		Language:     "swe",
		PartOfSpeech: "NN",
		WordParts:    "ap+nos",
		Transcriptions: []lex.Transcription{
			{
				Strn:     "Aa: p n u: s",
				Language: "swe",
			},
		},
	}

	vali.ValidateEntry(&e)
	result = e.EntryValidations

	expect = []lex.EntryValidation{
		{
			RuleName: "primary_stress",
			Level:    "Fatal",
			Message:  "[...]"},
		{
			RuleName: "syllabic",
			Level:    "Format",
			Message:  "[...]"},
		{
			RuleName: "symbolset",
			Level:    "Format",
			Message:  "[...]"},
		{
			RuleName: "MaxOneSyllabic",
			Level:    "Fatal",
			Message:  "[...]"},
	}

	if len(result) != len(expect) || result[0].RuleName != expect[0].RuleName || result[1].RuleName != expect[1].RuleName {
		t.Errorf(fsExp, expect, result)
	}
}
