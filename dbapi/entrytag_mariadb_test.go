package dbapi

import (
	"database/sql"
	"log"
	//"os"
	"testing"

	"github.com/stts-se/pronlex/lex"
)

func TestEntryTag1Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test11")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	if err != nil {

		t.Errorf("Failed to create lexicon db: %v", err)
	}

	l := lexicon{name: "entrytag_test", symbolSetName: "ZZ", locale: "ll"}
	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf("Ooops! : %v", err)
	}

	tx, err := db.Begin()

	defer tx.Commit()
	defer db.Close()
	if err != nil {
		t.Errorf("Failed to start transaction : %v", err)
	}

	// Insert tag for entry that doesn't exist
	err = mariaDBIF{}.insertEntryTagTx(tx, 0, "ohno", "homograph_1")
	//t.Errorf("Error : %v", err)
	if err == nil {
		t.Errorf("Expected error for nonexisting entry id, but got nil")
	}

	tx.Rollback()

	// Two different entris with the same orthography
	t1 := lex.Transcription{Strn: "A: p a", Language: "Svetsko"}
	t2 := lex.Transcription{Strn: "a pp a", Language: "svinspråket"}

	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "entrytag_1",
		Transcriptions: []lex.Transcription{t1, t2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	t1b := lex.Transcription{Strn: "A: p ' o", Language: "Svetsko"}
	t2b := lex.Transcription{Strn: "a p ' o", Language: "svinspråket"}

	e2 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "entrytag_2",
		Transcriptions: []lex.Transcription{t1b, t2b},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	_, err = mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1, e2})
	if err != nil {
		t.Errorf("failed to insert entries : %v", err)
	}

	q := Query{Words: []string{"apa"}, Page: 0, PageLength: 25}

	//var entries map[string][]lex.Entry
	entries, err := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q) // GetEntries(db, q)
	if err != nil {
		t.Errorf("Nooo! : %v", err)
	}
	if w, g := 1, len(entries); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
	}

	var ent1 lex.Entry
	var ent2 lex.Entry
	// We assume that the entry IDs are 1 and 2
	for _, e := range entries["apa"] {
		w1 := "entrytag_1"
		if e.ID == 1 && e.Tag != w1 {
			t.Errorf("Expected '%s' got '%s'", w1, e.Tag)
		}
		if e.ID == 1 {
			ent1 = e // Save for update test
		}

		w2 := "entrytag_2"
		if e.ID == 2 && e.Tag != w2 {
			t.Errorf("Expected '%s' got '%s'", w2, e.Tag)
		}
		if e.ID == 2 {
			ent2 = e // Save for update test
		}

	}

	// Change tag before update
	w := "entrytag_1b"
	ent1.Tag = w

	entUpdate, updated, err := mariaDBIF{}.updateEntry(db, ent1)
	if err != nil {
		t.Errorf("updateEntry failed : %v", err)
	}
	if !updated {
		t.Errorf("Expected entry to be updated, but nothing happened")
	}

	if entUpdate.Tag != w {
		t.Errorf("Wanted '%s' got '%s'", w, entUpdate.Tag)
	}

	// It should not be possible to assign the same Entry.Tag to two different entries
	ent2.Tag = w //No-no!
	_, updated2, err2 := mariaDBIF{}.updateEntry(db, ent2)
	if updated2 {
		t.Errorf("did not expect entry to be updated. disappointed.")
	}
	if err2 == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestEntryTag2Mariadb(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test12")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	if err != nil {

		t.Errorf("Failed to create lexicon db: %v", err)
	}

	l := lexicon{name: "entrytag_test", symbolSetName: "ZZ", locale: "ll"}
	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf("failed defineLexicon : %v", err)
	}

	tx, err := db.Begin()

	defer tx.Commit()
	defer db.Close()
	if err != nil {
		t.Errorf("Failed to start transaction : %v", err)
	}

	// Insert tag for entry that doesn't exist
	err = mariaDBIF{}.insertEntryTagTx(tx, 0, "ohno", "homograph2")
	//t.Errorf("Error : %v", err)
	if err == nil {
		t.Errorf("Expected error for nonexisting entry id, but got nil")
	}

	tx.Rollback()

	// Two different entris with the same orthography
	t1 := lex.Transcription{Strn: "A: p a", Language: "Svetsko"}
	t2 := lex.Transcription{Strn: "a pp a", Language: "svinspråket"}

	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "",
		Transcriptions: []lex.Transcription{t1, t2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	t1b := lex.Transcription{Strn: "A: p ' o", Language: "Svetsko"}
	t2b := lex.Transcription{Strn: "a p ' o", Language: "svinspråket"}

	e2 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "entrytag_2",
		Transcriptions: []lex.Transcription{t1b, t2b},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	_, err = mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1, e2})
	if err != nil {
		t.Errorf("failed to insert entries : %v", err)
	}

	// Test Query.TagLike

	q00 := Query{TagLike: "entrytag_2"}
	entries00, err00 := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q00)
	if err00 != nil {
		t.Errorf("Got error: %v", err00)
	}
	if w, g := 1, len(entries00); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
	}

	q := Query{Words: []string{"apa"}, Page: 0, PageLength: 25}

	//var entries map[string][]lex.Entry
	entries, err := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q) // GetEntries(db, q)
	if err != nil {
		t.Errorf("lookUpIntoMap : %v", err)
	}
	if w, g := 1, len(entries); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
	}

	var ent1 lex.Entry
	var ent2 lex.Entry
	// We assume that the entry IDs are 1 and 2
	for _, e := range entries["apa"] {
		w1 := ""
		if e.ID == 1 && e.Tag != w1 {
			t.Errorf("Expected '%s' got '%s'", w1, e.Tag)
		}
		if e.ID == 1 {
			ent1 = e // Save for update test
		}

		w2 := "entrytag_2"
		if e.ID == 2 && e.Tag != w2 {
			t.Errorf("Expected '%s' got '%s'", w2, e.Tag)
		}
		if e.ID == 2 {
			ent2 = e // Save for update test
		}

	}

	// Change tag before update
	w := "entrytag_1b"
	ent1.Tag = w

	entUpdate, updated, err := mariaDBIF{}.updateEntry(db, ent1)
	if err != nil {
		t.Errorf("updateEntry failed : %v", err)
	}
	if !updated {
		t.Errorf("Expected entry to be updated, but nothing happened")
	}

	if entUpdate.Tag != w {
		t.Errorf("Wanted '%s' got '%s'", w, entUpdate.Tag)
	}

	// It should not be possible to assign the same Entry.Tag to two different entries
	ent2.Tag = w //No-no!
	_, updated2, err2 := mariaDBIF{}.updateEntry(db, ent2)
	if updated2 {
		t.Errorf("did not expect entry to be updated. disappointed.")
	}
	if err2 == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestMultipleTags1MariaDB(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test12")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	if err != nil {

		t.Errorf("Failed to create lexicon db: %v", err)
	}

	l := lexicon{name: "multipletags_test", symbolSetName: "ZZ", locale: "ll"}
	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf("failed defineLexicon : %v", err)
	}

	tx, err := db.Begin()

	defer tx.Commit()
	defer db.Close()
	if err != nil {
		t.Errorf("Failed to start transaction : %v", err)
	}

	if err != nil {
		t.Errorf("Failed to start transaction : %v", err)
	}

	// Insert tag for entry that doesn't exist
	err = mariaDBIF{}.insertEntryTagTx(tx, 0, "ohno", "homograph3")
	//t.Errorf("Error : %v", err)
	if err == nil {
		t.Errorf("Expected error for nonexisting entry id, but got nil")
	}

	tx.Rollback()

	// Two different entris with the same orthography
	t1 := lex.Transcription{Strn: "A: p a", Language: "Svetsko"}
	t2 := lex.Transcription{Strn: "a pp a", Language: "svinspråket"}

	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "entrytag_1",
		Transcriptions: []lex.Transcription{t1, t2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	t1b := lex.Transcription{Strn: "A: p ' o", Language: "Svetsko"}
	t2b := lex.Transcription{Strn: "a p ' o", Language: "svinspråket"}

	e2 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "entrytag_2",
		Transcriptions: []lex.Transcription{t1b, t2b},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}
	e3 := lex.Entry{Strn: "apan",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apan",
		Language:       "XYZZ",
		Preferred:      true,
		Tag:            "entrytag_3",
		Transcriptions: []lex.Transcription{t1b, t2b},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"}}

	_, err = mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1, e2, e3})
	if err != nil {
		t.Errorf("failed to insert entries : %v", err)
	}

	// Test Query.MultipleTags
	qMulti := Query{MultipleTags: true}
	entries, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, qMulti)
	if err != nil {
		t.Errorf("Got error: %v", err)
		return
	}
	if w, g := 2, len(entries); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
		return
	}

	return

	// remove one of the tags
	entries[0].Tag = ""

	entUpdate, updated, err := mariaDBIF{}.updateEntry(db, entries[0])
	if err != nil {
		t.Errorf("updateEntry failed : %v", err)
	}
	if !updated {
		t.Errorf("Expected entry to be updated, but nothing happened")
	}

	if entUpdate.Tag != "" {
		t.Errorf("Wanted '%s' got '%s'", "", entUpdate.Tag)
	}

	// Test Query.MultipleTags
	entries2, err := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, qMulti)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if w, g := 0, len(entries2); w != g {
		t.Errorf("Expected '%v' got '%#v'", []lex.Entry{}, entries2)
	}

}
