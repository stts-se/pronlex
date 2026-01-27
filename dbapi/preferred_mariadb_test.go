package dbapi

import (
	"database/sql"
	"log"
	"testing"

	"github.com/stts-se/pronlex/lex"
)

func TestPreferred1MariaDB(t *testing.T) {
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

	l := lexicon{name: "preferred_test", symbolSetName: "ZZ", locale: "ll"}
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

	t1 := "city"
	t2 := "drink"
	e1 := lex.Entry{Strn: "rom",
		PartOfSpeech:   "PM",
		WordParts:      "rom",
		Language:       "",
		Preferred:      true,
		Tag:            t1,
		Transcriptions: []lex.Transcription{{Strn: "\" r u m", Language: ""}},
		EntryStatus:    lex.EntryStatus{Name: "unchecked", Source: "imported"}}

	e2 := lex.Entry{Strn: "rom",
		PartOfSpeech:   "NN",
		WordParts:      "rom",
		Language:       "",
		Preferred:      false,
		Tag:            t2,
		Transcriptions: []lex.Transcription{{Strn: "\" r o m", Language: ""}},
		EntryStatus:    lex.EntryStatus{Name: "unchecked", Source: "imported"}}

	// Insert entries
	_, err = mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1, e2})
	if err != nil {
		t.Errorf("failed to insert entries : %v", err)
	}

	// Fetch inserted entries
	q := Query{Words: []string{"rom"}, Page: 0, PageLength: 25}

	entries, err := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf("Nooo! : %v", err)
	}
	if w, g := 1, len(entries); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
	}

	var ent1 lex.Entry
	var ent2 lex.Entry
	seenTag1 := false
	seenTag2 := false
	for _, e := range entries["rom"] {
		if e.Tag == t1 {
			ent1 = e // Save for update test
			seenTag1 = true
		}
		if e.Tag == t2 {
			ent2 = e // Save for update test
			seenTag2 = true
		}
	}
	if !seenTag1 {
		t.Errorf("Couldn't find tag %s in looked up entries %#v", t1, entries)
	}
	if !seenTag2 {
		t.Errorf("Couldn't find tag %s in looked up entries %#v", t2, entries)
	}

	// Verify correct initial pref tags
	if !ent1.Preferred {
		t.Errorf("Expected preferred:true for ent1, found %#v", ent1)
	}
	if ent2.Preferred {
		t.Errorf("Expected preferred:false for ent2, found %#v", ent2)
	}

	// Set language
	ent1.Language = "sv"
	ent1.EntryStatus = lex.EntryStatus{Name: "ok", Source: "hanna"}

	// Update entry with new language
	entUpdate, updated, err := mariaDBIF{}.updateEntry(db, ent1)
	if err != nil {
		t.Errorf("updateEntry failed : %v", err)
	}
	if !updated {
		t.Errorf("Expected entry to be updated, but nothing happened")
	}
	if !entUpdate.Preferred {
		t.Errorf("Expected updated entry have preferred tag true, but found %#v", entUpdate)
	}

	// Fetch entries again
	q = Query{Words: []string{"rom"}, Page: 0, PageLength: 25}

	entries, err = mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf("Nooo! : %v", err)
	}
	if w, g := 1, len(entries); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
	}

	seenTag1 = false
	seenTag2 = false
	for _, e := range entries["rom"] {
		if e.Tag == t1 {
			ent1 = e // Save for update test
			seenTag1 = true
		}
		if e.Tag == t2 {
			ent2 = e // Save for update test
			seenTag2 = true
		}
	}
	// Verify that we still have correct pref tags
	if !ent1.Preferred {
		t.Errorf("Expected preferred:true for ent1, found %#v", ent1)
	}
	if ent2.Preferred {
		t.Errorf("Expected preferred:false for ent2, found %#v", ent2)
	}

	// Change preferred
	ent2.Preferred = true

	// Update entry with new language
	entUpdate, updated, err = mariaDBIF{}.updateEntry(db, ent2)
	if err != nil {
		t.Errorf("updateEntry failed : %v", err)
	}
	if !updated {
		t.Errorf("Expected entry to be updated, but nothing happened")
	}
	if !entUpdate.Preferred {
		t.Errorf("Expected updated entry have preferred tag true, but found %#v", entUpdate)
	}

	// Fetch entries again
	q = Query{Words: []string{"rom"}, Page: 0, PageLength: 25}

	entries, err = mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf("Nooo! : %v", err)
	}
	if w, g := 1, len(entries); w != g {
		t.Errorf("Expected '%d' got '%d'", w, g)
	}

	seenTag1 = false
	seenTag2 = false
	for _, e := range entries["rom"] {
		if e.Tag == t1 {
			ent1 = e // Save for update test
			seenTag1 = true
		}
		if e.Tag == t2 {
			ent2 = e // Save for update test
			seenTag2 = true
		}
	}
	// Verify that we have new corrected pref tags
	if ent1.Preferred {
		t.Errorf("Expected preferred:false for ent1, found %#v", ent1)
	}
	if !ent2.Preferred {
		t.Errorf("Expected preferred:true for ent2, found %#v", ent2)
	}

}
