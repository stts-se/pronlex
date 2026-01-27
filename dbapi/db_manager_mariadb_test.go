package dbapi

import (
	"database/sql"
	"fmt"
	"log"
	//"os"
	"testing"

	"github.com/stts-se/pronlex/lex"
)

func Test_DBManagerMariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	// dbPath1 := "./testlex_listlex1.db"
	// dbPath2 := "./testlex_listlex2.db"

	// if _, err := os.Stat(dbPath1); !os.IsNotExist(err) {
	// 	err := os.Remove(dbPath1)

	// 	if err != nil {
	// 		log.Fatalf("failed to remove '%s' : %v", dbPath1, err)
	// 	}
	// }

	// if _, err := os.Stat(dbPath2); !os.IsNotExist(err) {
	// 	err := os.Remove(dbPath2)

	// 	if err != nil {
	// 		log.Fatalf("failed to remove '%s' : %v", dbPath2, err)
	// 	}
	// }

	db1, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test8")
	if err != nil {
		log.Fatal(err)
	}

	//defer db.Close()

	//db1, err := sql.Open("sqlite3_with_regexp", dbPath1)
	if err != nil {
		log.Fatal(err)
	}

	db2, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test9")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer db.Close()

	//db2, err := sql.Open("sqlite3_with_regexp", dbPath2)
	if err != nil {
		log.Fatal(err)
	}

	// _, err = db1.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db2.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db1.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)
	// _, err = db2.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	defer db1.Close()
	defer db2.Close()

	_, err = execSchemaMariadb(db1) // Creates new lexicon database
	if err != nil {
		log.Fatalf("NO! creating db1 for mariadb failed: %v", err)
	}
	_, err = execSchemaMariadb(db2) // Creates new lexicon database
	if err != nil {
		log.Fatalf("NO! creating db2 for mariadb failed: %v", err)
	}

	dbm := NewMariaDBManager()
	dbm.AddDB("db1", db1)
	dbm.AddDB("db2", db2)

	l1_1 := lex.LexName("zuperlex1")
	l1_2 := lex.LexName("zuperlex2")
	l1_3 := lex.LexName("zuperlex3")

	l2_1 := lex.LexName("zuperlex1")
	l2_2 := lex.LexName("zuperlex2")
	l2_3 := lex.LexName("zuperduperlex")

	err = dbm.DefineLexicons(lex.DBRef("db1"), "sv_sampa", "sv", l1_1, l1_2, l1_3)
	if err != nil {
		t.Errorf("Quack! %v", err)
	}
	err = dbm.DefineLexicons(lex.DBRef("db2"), "sv_sampa", "sv_SE", l2_1, l2_2, l2_3)
	if err != nil {
		t.Errorf("Quack! %v", err)
	}

	lexs, err := dbm.ListLexicons()
	if err != nil {
		t.Errorf("Quack! %v", err)
	}

	if w, g := 6, len(lexs); w != g {
		t.Errorf("wanted %v got %v", w, g)
	}

	lexsM := make(map[lex.LexRef]bool)
	for _, l := range lexs {
		lexsM[l.LexRef] = true
	}

	if w := lex.NewLexRef("db1", "zuperlex1"); !lexsM[w] {
		t.Errorf("expected db not found: '%s'", w)
	}
	if w := lex.NewLexRef("db1", "zuperlex2"); !lexsM[w] {
		t.Errorf("expected db not found: '%s'", w)
	}

	if w := lex.NewLexRef("db1", "zuperlex3"); !lexsM[w] {
		t.Errorf("expected db not found: '%s'", w)
	}

	if w := lex.NewLexRef("db2", "zuperlex1"); !lexsM[w] {
		t.Errorf("expected db not found: '%s'", w)
	}
	if w := lex.NewLexRef("db2", "zuperlex2"); !lexsM[w] {
		t.Errorf("expected db not found: '%s'", w)
	}

	if w := lex.NewLexRef("db2", "zuperduperlex"); !lexsM[w] {
		t.Errorf("expected db not found: '%s'", w)
	}

	//e1 := lex.Entry{Strn: "hus", Transcriptions: []lex.Transcription{lex.Transcription{Strn: `" h u: s`}}}
	t1 := lex.Transcription{Strn: "A: p a", Language: "Svetsko"}
	t2 := lex.Transcription{Strn: "a pp a", Language: "svinspråket"}

	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Transcriptions: []lex.Transcription{t1, t2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	ids, err := dbm.InsertEntries(lex.NewLexRef("db2", "zuperduperlex"), []lex.Entry{e1})
	if w, g := 1, len(ids); w != g {
		t.Errorf("Wanted %v got %v", w, g)
	}
	if err != nil {
		t.Errorf("dbm.InsertEntries: %v", err)
	}

	q := Query{Words: []string{"apa"}}
	lookRes, err := dbm.LookUpIntoMap(DBMQuery{[]lex.LexRef{lex.NewLexRef("db2", "zuperduperlex")}, q})
	if err != nil {
		t.Errorf("dbm.LookUpIntoMap : %v", err)
	}
	if w, g := 1, len(lookRes); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	ents := lookRes["db2"]
	//fmt.Printf("%v\n", ents)
	if w, g := 1, len(ents); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	ids, err = dbm.InsertEntries(lex.NewLexRef("db1", "zuperlex1"), []lex.Entry{e1})
	if w, g := 1, len(ids); w != g {
		t.Errorf("Wanted %v got %v", w, g)
	}
	if err != nil {
		t.Errorf("dbm.InsertEntries: %v", err)
	}

	lookRes, err = dbm.LookUpIntoMap(DBMQuery{[]lex.LexRef{lex.NewLexRef("db2", "zuperduperlex"), lex.NewLexRef("db1", "zuperlex1")}, q})

	if w, g := 2, len(lookRes); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	if err != nil {
		t.Errorf("dbm.InsertEntries: %v", err)
	}
	ents = lookRes["db2"]
	//fmt.Printf("%v\n", ents)
	if w, g := 1, len(ents); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	ents = lookRes["db1"]
	//fmt.Printf("%v\n", ents)
	if w, g := 1, len(ents); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	t2_1 := lex.Transcription{Strn: "u: p a", Language: "Svetsko"}
	t2_2 := lex.Transcription{Strn: "u pp a", Language: "svinspråket"}

	e2 := lex.Entry{Strn: "upa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Transcriptions: []lex.Transcription{t2_1, t2_2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}
	e3 := lex.Entry{Strn: "uppa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Transcriptions: []lex.Transcription{t2_1, t2_2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	idz, err := dbm.InsertEntries(lex.NewLexRef("db1", "zuperlex3"), []lex.Entry{e2, e3})
	if len(idz) != 2 {
		t.Errorf("Freaky!")
	}
	if err != nil {
		t.Errorf("gah! : %v", err)
	}

	lookRez, err := dbm.LookUpIntoMap(DBMQuery{[]lex.LexRef{lex.NewLexRef("db2", "zuperduperlex"), lex.NewLexRef("db1", "zuperlex1"), lex.NewLexRef("db1", "zuperlex3")}, Query{WordRegexp: "."}})
	//fmt.Printf("%v\n", lookRez)
	if err != nil {
		t.Errorf("geh! : %v", err)
	}
	if w, g := 2, len(lookRez); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 3, len(lookRez["db1"]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	if w, g := 1, len(lookRez["db2"]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	// Update a DB entry
	lookUpApa, err := dbm.LookUpIntoMap(DBMQuery{[]lex.LexRef{lex.NewLexRef("db1", "zuperlex1")}, Query{Words: []string{"apa"}}})
	if err != nil {
		t.Errorf("LookUp failed : %v", err)
	}

	apaE := lookUpApa["db1"][0]
	w1 := "zzzu: p a"
	w2 := "Svetzzz"
	apaE.Transcriptions = []lex.Transcription{{Strn: w1, Language: w2}}

	//_, _, err = dbm.UpdateEntry(lex.DBRef("db1"), apaE)
	_, _, err = dbm.UpdateEntry(apaE)
	if err != nil {
		t.Errorf("serious! : %v", err)
	}

	lookUpApa, err = dbm.LookUpIntoMap(DBMQuery{[]lex.LexRef{lex.NewLexRef("db1", "zuperlex1")}, Query{Words: []string{"apa"}}})
	if err != nil {
		t.Errorf("LookUp failed : %v", err)
	}

	apaE = lookUpApa["db1"][0]
	if w, g := 1, len(apaE.Transcriptions); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	if w, g := w1, apaE.Transcriptions[0].Strn; w != g {
		t.Errorf("wanted %s got %s", w, g)
	}
	if w, g := w2, apaE.Transcriptions[0].Language; w != g {
		t.Errorf("wanted %s got %s", w, g)
	}

	fmt.Printf("")
	//fmt.Printf("%v\n", lexs)
}
