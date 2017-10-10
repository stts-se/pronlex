package dbapi

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stts-se/pronlex/lex"
)

func TestEntryTag(t *testing.T) {

	dbPath := "./testlex_entrytag.db"
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		err := os.Remove(dbPath)
		if err != nil {
			t.Errorf("failed to remove %s : %v", dbPath, err)
		}
	}

	db, err := sql.Open("sqlite3_with_regexp", dbPath)
	if err != nil {
		t.Errorf("Failed to open db file %s : %v", dbPath, err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Errorf("Failed to call PRAGMA on db : %v", err)
	}
	_, err = db.Exec("PRAGMA case_sensitive_like=ON")
	if err != nil {
		t.Errorf("Failed to exec PRAGMA call %v", err)
	}
	defer db.Close()

	_, err = execSchema(db) // Creates new lexicon database
	if err != nil {

		t.Errorf("Failed to create lexicon db: %v", err)
	}

	l := lexicon{name: "entrytag_test", symbolSetName: "ZZ"}
	l, err = defineLexicon(db, l)

	tx, err := db.Begin()

	defer tx.Commit()
	defer db.Close()
	if err != nil {
		t.Errorf("Failed to start transaction : %v", err)
	}

	// Insert tag for entry that doesn't exist
	err = insertEntryTagTx(tx, 0, "ohno")
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

	_, err = insertEntries(db, l, []lex.Entry{e1, e2})
	if err != nil {
		t.Errorf("failed to insert entries : %v", err)
	}

	q := Query{Words: []string{"apa"}, Page: 0, PageLength: 25}

	//var entries map[string][]lex.Entry
	entries, err := lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q) // GetEntries(db, q)
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

	entUpdate, updated, err := updateEntry(db, ent1)
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
	_, updated2, err2 := updateEntry(db, ent2)
	if updated2 {
		t.Errorf("did not expect entry to be updated. disappointed.")
	}
	if err2 == nil {
		t.Errorf("Expected error, got nil")
	}
}
