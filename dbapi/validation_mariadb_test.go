package dbapi

import (
	"database/sql"

	"log"
	//"os"
	"testing"

	"reflect"

	"github.com/dlclark/regexp2"
	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/validation"
	"github.com/stts-se/pronlex/validation/rules"
	"github.com/stts-se/symbolset"
)

//var vfs = "Wanted: '%v' got: '%v'"

func createValidatorMariadbTest() validation.Validator {
	name := "sampa"
	symbols := []symbolset.Symbol{
		{String: "a", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "A:", Cat: symbolset.Syllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "b", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "p", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "N", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "n", Cat: symbolset.NonSyllabic, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: " ", Cat: symbolset.PhonemeDelimiter, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: ".", Cat: symbolset.SyllableDelimiter, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\"", Cat: symbolset.Stress, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
		{String: "\"\"", Cat: symbolset.Stress, Desc: "", IPA: symbolset.IPASymbol{String: "", Unicode: ""}},
	}
	ss, err := symbolset.NewSymbolSet(name, symbols)
	ff("failed to init symbols : %v", err)

	primaryStressRe, err := rules.ProcessTransRe(ss, "\"")
	ff("%v", err)

	syllabicRe, err := rules.ProcessTransRe(ss, "^(\"\"|\"|%)? *(nonsyllabic +)*syllabic( +nonsyllabic)*( (.|-) (\"\"|\"|%)? *(nonsyllabic +)*syllabic( +nonsyllabic)*)*$")
	ff("%v", err)

	reFrom, err := regexp2.Compile("(.)\\1[+]\\1", regexp2.None)
	ff("%v", err)

	decomp2Orth := rules.Decomp2Orth{CompDelim: "+",
		AcceptEmptyDecomp: true,
		PreFilterWordPartString: func(s string) (string, error) {
			res, err := reFrom.Replace(s, "$1+$1", 0, -1)
			if err != nil {
				return s, err
			}
			return res, nil
		}}

	repeatedPhnRe, err := rules.ProcessTransRe(ss, "symbol( +[.~])? +\\1")
	ff("%v", err)

	var v = validation.Validator{
		Name: ss.Name,
		Rules: []validation.Rule{
			rules.MustHaveTrans{},
			rules.NoEmptyTrans{},
			rules.RequiredTransRe{
				NameStr:  "primary_stress",
				LevelStr: "Fatal",
				Message:  "Primary stress required",
				Re:       primaryStressRe,
			},
			rules.RequiredTransRe{
				NameStr:  "syllabic",
				LevelStr: "Format",
				Message:  "Each syllable needs a syllabic phoneme",
				Re:       syllabicRe,
			},
			rules.IllegalTransRe{
				NameStr:  "repeated_phonemes",
				LevelStr: "Fatal",
				Message:  "Repeated phonemes cannot be used within the same morpheme",
				Re:       repeatedPhnRe,
			},
			decomp2Orth,
			rules.SymbolSetRule{
				SymbolSet: ss,
			},
		}}
	return v
}

func vInsertEntriesMariadb(t *testing.T, lexName string) (*sql.DB, string) {

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test13")
	if err != nil {
		log.Fatal(err)
	}

	//defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	ff("Failed to create lexicon db: %v", err)

	l := lexicon{name: lexName, symbolSetName: "ZZ", locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(vfs, nil, err)
	}

	lxs, err := mariaDBIF{}.listLexicons(db)
	if err != nil {
		t.Errorf(vfs, nil, err)
	}
	if len(lxs) != 1 {
		t.Errorf(vfs, 1, len(lxs))
	}

	t1a := lex.Transcription{Strn: "\" A: p a", Language: "sv-se"}
	t1b := lex.Transcription{Strn: "\" a p a", Language: "sv-se"}

	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Transcriptions: []lex.Transcription{t1a, t1b},
		EntryStatus:    lex.EntryStatus{Name: "old", Source: "tst"}}

	t2a := lex.Transcription{Strn: "\" A: p a n", Language: "sv-se"}
	t2b := lex.Transcription{Strn: "A: p a n", Language: "sv-se"}

	e2 := lex.Entry{Strn: "apan",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apan",
		Language:       "XYZZ",
		Transcriptions: []lex.Transcription{t2a, t2b},
		EntryStatus:    lex.EntryStatus{Name: "old", Source: "tst"}}

	t3a := lex.Transcription{Strn: "\" A . p a n", Language: "sv-se"}
	e3 := lex.Entry{Strn: "appan",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "appan",
		Language:       "XYZZ",
		Transcriptions: []lex.Transcription{t3a},
		EntryStatus:    lex.EntryStatus{Name: "old", Source: "tst"}}

	t4a := lex.Transcription{Strn: "\" A . p a n", Language: "sv-se"}
	e4 := lex.Entry{Strn: "appan",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "appan",
		Language:       "XYZZ",
		Transcriptions: []lex.Transcription{t4a},
		EntryStatus:    lex.EntryStatus{Name: "old", Source: "tst"}}

	_, errx := mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1, e2, e3, e4})
	if errx != nil {
		t.Errorf(vfs, "nil", errx)
	}
	return db, l.name
}

func Test_Validation1Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	db, lexName := vInsertEntriesMariadb(t, "test1")
	v := createValidatorMariadbTest()

	q := Query{}

	stats, err := validate(mariaDBIF{}, db, []lex.LexName{lex.LexName("test1")}, SilentLogger{}, v, q)
	ff("validation failed : %v", err)

	expect := ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expect, stats) {
		t.Errorf(vfs, expect, stats)
	}

	// check stats saved in db
	lexStats, err := mariaDBIF{}.validationStats(db, lexName)
	ff("validation stats failed : %v", err)

	expectFull := ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expectFull, lexStats) {
		t.Errorf(vfs, expectFull, lexStats)
	}
}

func Test_Validation2Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	db, lexName := vInsertEntriesMariadb(t, "test2")
	v := createValidatorMariadbTest()

	q := Query{WordRegexp: "a$"}

	// test 1
	stats, err := validate(mariaDBIF{}, db, []lex.LexName{lex.LexName("test2")}, SilentLogger{}, v, q)
	ff("validation failed : %v", err)

	expect := ValStats{
		TotalEntries:     1,
		ValidatedEntries: 1,
		TotalValidations: 0,
		InvalidEntries:   0,
		Levels:           make(map[string]int),
		Rules:            make(map[string]int),
	}
	if !reflect.DeepEqual(expect, stats) {
		t.Errorf(vfs, expect, stats)
	}

	// check stats saved in db
	lexStats, err := mariaDBIF{}.validationStats(db, lexName)
	ff("validation stats failed : %v", err)

	expectFull := ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 0,
		InvalidEntries:   0,
		Levels:           make(map[string]int),
		Rules:            make(map[string]int),
	}

	if !reflect.DeepEqual(expectFull, lexStats) {
		t.Errorf(vfs, expectFull, lexStats)
	}

	// test 2
	q = Query{}

	stats, err = validate(mariaDBIF{}, db, []lex.LexName{lex.LexName("test2")}, SilentLogger{}, v, q)
	ff("validation failed : %v", err)

	expect = ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expect, stats) {
		t.Errorf(vfs, expect, stats)
	}

	// check stats saved in db
	lexStats, err = mariaDBIF{}.validationStats(db, lexName)
	ff("validation stats failed : %v", err)

	expectFull = ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expectFull, lexStats) {
		t.Errorf(vfs, expectFull, lexStats)
	}
}

func Test_ValidationUpdate1Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	db, lexName := vInsertEntriesMariadb(t, "test3")
	v := createValidatorMariadbTest()
	ew := lex.EntrySliceWriter{}
	err := mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName("test3")}, Query{WordLike: "%"}, &ew)
	ff("lookup failed : %v", err)

	for _, e := range ew.Entries {
		v.ValidateEntry(&e)
		err = mariaDBIF{}.updateValidation(db, []lex.Entry{e})
		ff("update validation failed : %v", err)
	}

	stats, err := mariaDBIF{}.validationStats(db, lexName)
	ff("validation stats failed : %v", err)

	expect := ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expect, stats) {
		t.Errorf(vfs, expect, stats)
	}
}

func Test_ValidationUpdate2Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	db, lexName := vInsertEntriesMariadb(t, "test4")
	v := createValidatorMariadbTest()
	ew := lex.EntrySliceWriter{}
	err := mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName("test4")}, Query{WordLike: "%"}, &ew)
	ff("lookup failed : %v", err)

	var es []lex.Entry
	for _, e := range ew.Entries {
		v.ValidateEntry(&e)
		es = append(es, e)
	}
	err = mariaDBIF{}.updateValidation(db, es)
	ff("update validation failed : %v", err)

	stats, err := mariaDBIF{}.validationStats(db, lexName)
	ff("validation stats failed : %v", err)

	expect := ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expect, stats) {
		t.Errorf(vfs, expect, stats)
	}
}

func Test_ValidationUpdate3Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	db, lexName := vInsertEntriesMariadb(t, "test5")
	v := createValidatorMariadbTest()
	ew := lex.EntrySliceWriter{}
	err := mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName("test5")}, Query{WordLike: "%"}, &ew)
	ff("lookup failed : %v", err)

	es, _ := v.ValidateEntries(ew.Entries)
	err = mariaDBIF{}.updateValidation(db, es)
	ff("update validation failed : %v", err)

	stats, err := mariaDBIF{}.validationStats(db, lexName)
	ff("validation stats failed : %v", err)

	expect := ValStats{
		TotalEntries:     4,
		ValidatedEntries: 4,
		TotalValidations: 5,
		InvalidEntries:   3,
		Levels: map[string]int{
			"fatal":  3,
			"format": 2,
		},
		Rules: map[string]int{
			"symbolset (fatal)":      2,
			"primary_stress (fatal)": 1,
			"syllabic (format)":      2,
		},
	}

	if !reflect.DeepEqual(expect, stats) {
		t.Errorf(vfs, expect, stats)
	}
}
