package dbapi

import (
	"database/sql"
	"flag"
	//"fmt"
	"reflect"
	"time"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/validation"
	"github.com/stts-se/symbolset"
	//"github.com/mattn/go-sqlite3"
	"log"
	//"os"
	//"regexp"
	"testing"
)

// Set up for local testing, run:
// $ sudo mysql -u root < scripts/mariadb_setup.sql

var WithMariaDB = flag.Bool("mariadb", false, "Use this flag to inhibit MariaDB tests")

// ff is a place holder to be replaced by proper error handling
func ff(f string, err error) {
	if err != nil {
		log.Fatalf(f, err)
	}
}

func execSchemaMariadb(db *sql.DB) (sql.Result, error) {
	ti := time.Now()

	var err error
	var res sql.Result

	res, err = db.Exec(mariaDBDropTableStmt)
	if err != nil {
		return res, err
	}

	for _, s := range MariaDBSchema {

		res, err := db.Exec(s)
		if err != nil {
			return res, err
		}

	}
	_ = ti
	//fmt.Printf("[dbapi_test] db.Exec(Schema) took %v\n", time.Since(ti))
	return res, err
}

func Test_insertEntries(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test1")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database

	ff("Failed to create lexicon db: %v", err)

	// TODO Borde returnera error
	//CreateTables(db, cmds)

	l := lexicon{name: "test", symbolSetName: "ZZ", locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	lxs, err := mariaDBIF{}.listLexicons(db)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(lxs) != 1 {
		t.Errorf(fs, 1, len(lxs))
	}
	if lxs[0].name != "test" {
		t.Errorf(fs, "test", lxs[0].name)
	}
	if lxs[0].id <= 0 {
		t.Errorf(fs, ">0", lxs[0].id)
	}
	if lxs[0].symbolSetName != "ZZ" {
		t.Errorf(fs, "ZZ", lxs[0].symbolSetName)
	}

	lx, err := mariaDBIF{}.getLexicon(db, "test")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if w, g := "test", lx.name; w != g {
		t.Errorf("Wanted %s got %s", w, g)
	}
	if w, g := "ZZ", lx.symbolSetName; w != g {
		t.Errorf("Wanted %s got %s", w, g)
	}
	lx, err = mariaDBIF{}.getLexicon(db, "xyzzhga_skdjdj")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if w, g := "", lx.name; w != g {
		t.Errorf("Wanted empty string, got '%s'", g)
	}

	t1 := lex.Transcription{Strn: "A: p a", Language: "Svetsko"}
	t2 := lex.Transcription{Strn: "a pp a", Language: "svinspråket"}

	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech:   "NN",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Transcriptions: []lex.Transcription{t1, t2},
		EntryStatus:    lex.EntryStatus{Name: "old1", Source: "tst"},
	}

	_, errx := mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1})
	if errx != nil {
		t.Errorf(fs, "nil", errx)
		return
	}

	//time.Sleep(2000 * time.Millisecond)

	// Check that there are things in db:
	q := Query{Words: []string{"apa"}, Page: 0, PageLength: 25}

	var entries map[string][]lex.Entry
	entries, err = mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q) // GetEntries(db, q)

	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if got, want := len(entries), 1; got != want {
		t.Errorf(fs, got, want)
	}

	ea := entries["apa"][0]
	if got, want := ea.Morphology, "NEU UTR"; got != want {
		t.Errorf(fs, got, want)
	}
	if got, want := ea.Preferred, true; got != want {
		t.Errorf(fs, got, want)
	}

	for _, e := range entries {
		ts := len(e[0].Transcriptions)
		if ts != 2 {
			t.Errorf(fs, 2, ts)
		}
	}

	le := lex.Lemma{Strn: "apa", Reading: "67t", Paradigm: "7(c)"}
	tx0, err := db.Begin()
	defer tx0.Commit()
	ff("transaction failed : %v", err)
	le2, err := mariaDBIF{}.insertLemma(tx0, le)
	if err != nil {
		t.Errorf("insertLemma : %v", err)
	}
	tx0.Commit()
	if le2.ID < 1 {
		t.Errorf(fs, "more than zero", le2.ID)
	}

	que := Query{TranscriptionLike: "%pp%"}
	var queRez lex.EntrySliceWriter
	err = mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName(l.name)}, que, &queRez)
	if err != nil {
		t.Errorf("Wanted nil, got %v", err)
	}
	if got, want := len(queRez.Entries), 1; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	tx00, err := db.Begin()
	ff("tx failed : %v", err)
	defer tx00.Commit()

	le3, err := mariaDBIF{}.setOrGetLemma(tx00, "apa", "67t", "7(c)")
	if err != nil {
		t.Errorf("setOrGetLemma : %v", err)
	}

	if le3.ID < 1 {
		t.Errorf(fs, "more than zero", le3.ID)
	}
	tx00.Commit()

	tx01, err := db.Begin()
	ff("tx failed : %v", err)
	defer tx01.Commit()
	err = mariaDBIF{}.associateLemma2Entry(tx01, le3, entries["apa"][0])
	if err != nil {
		t.Error(fs, nil, err)
	}
	tx01.Commit()

	//ess, err := GetEntries(db, q)
	//var esw lex.EntrySliceWriter
	ess, err := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	if len(ess) != 1 {
		t.Error("ERRRRRRROR")
	}
	lm := ess["apa"][0].Lemma
	if lm.ID < 1 {
		t.Errorf(fs, "id larger than zero", lm.ID)
	}

	if lm.Strn != "apa" {
		t.Errorf(fs, "apa", lm.Strn)
	}
	if lm.Reading != "67t" {
		t.Errorf(fs, "67t", lm.Reading)
	}

	//ees := GetEntriesFromIDs(db, []int64{ess["apa"][0].ID})
	ees, err := mariaDBIF{}.lookUpIntoMap(db, []lex.LexName{lex.LexName(l.name)}, Query{EntryIDs: []int64{ess["apa"][0].ID}})
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(ees) != 1 {
		t.Errorf(fs, 1, len(ees))
	}

	// Check that no entries with entryvalidation exist
	noev, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{HasEntryValidation: true})
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if got, want := len(noev), 0; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	// Change transcriptions and update db
	ees0 := ees["apa"][0]
	t10 := lex.Transcription{Strn: "A: p A:", Language: "Apo"}
	t10.AddSource("orangu1")
	t20 := lex.Transcription{Strn: "a p a", Language: "Sweinsprach"}
	t20.AddSource("orangu2")
	t30 := lex.Transcription{Strn: "a pp a", Language: "Mysko"}
	t30.AddSource("orangu3")
	t30.AddSource("orangu4")
	ees0.Transcriptions = []lex.Transcription{t10, t20, t30}
	// add new lex.EntryStatus
	ees0.EntryStatus = lex.EntryStatus{Name: "new", Source: "tst"}
	// new validation
	ees0.EntryValidations = []lex.EntryValidation{{Level: "severe", RuleName: "barf", Message: "it hurts"}}

	ees0.PartOfSpeech = "PM"
	ees0.Morphology = "F"
	ees0.Tag = "accent II"

	//time.Sleep(2000 * time.Millisecond)

	newE, updated, err := mariaDBIF{}.updateEntry(db, ees0)

	oldEntryStatus := ees0.EntryStatus
	newEntryStatus := newE.EntryStatus

	// Assert that the statuses have different time stamps
	if oldEntryStatus.Timestamp == newEntryStatus.Timestamp {
		t.Errorf("Expected different EntryStatus.Timestamp, got same: %#v\n", oldEntryStatus)
	}

	if err != nil {
		t.Errorf(fs, nil, err)
	}

	if !updated {
		t.Errorf(fs, true, updated)
	}

	if want, got := true, newE.Strn == ees0.Strn; !got {
		t.Errorf(fs, got, want)
	}

	eApa, err := mariaDBIF{}.getEntryFromID(db, ees0.ID)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(eApa.Transcriptions) != 3 {
		t.Errorf(fs, 3, len(eApa.Transcriptions))
	}

	if got, want := eApa.Transcriptions[0].Sources[0], "orangu1"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := len(eApa.Transcriptions[2].Sources), 2; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	if got, want := eApa.Transcriptions[2].Sources[0], "orangu4"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := eApa.EntryStatus.Name, "new"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := len(eApa.EntryValidations), 1; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	if got, want := eApa.EntryValidations[0].Level, "severe"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := eApa.EntryValidations[0].RuleName, "barf"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	if got, want := eApa.EntryValidations[0].Message, "it hurts"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := eApa.PartOfSpeech, "PM"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := eApa.Morphology, "F"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	if got, want := eApa.Tag, "accent ii"; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	// Check that one entry with entryvalidation exists
	noev, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{HasEntryValidation: true})
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if got, want := len(noev), 1; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	eApa.Lemma.Strn = "tjubba"
	eApa.WordParts = "fin+krog"
	eApa.Language = "gummiapa"
	eApa.EntryValidations = []lex.EntryValidation{}

	//
	c1 := lex.EntryComment{Label: "label1", Source: "secret", Comment: "strålande"}
	c2 := lex.EntryComment{Label: "label2", Source: "hämligt", Comment: "super super hemligt |)("}
	cmts := []lex.EntryComment{c1, c2}
	eApa.Comments = cmts

	//time.Sleep(2000 * time.Millisecond)
	newE2, updated, err := mariaDBIF{}.updateEntry(db, eApa)
	if err != nil {
		t.Errorf(fs, "nil", err)
	}
	if !updated {
		t.Errorf(fs, true, updated)
	}
	if want, got := true, newE2.Strn == eApa.Strn; !got {
		t.Errorf(fs, got, want)
	}

	eApax, err := mariaDBIF{}.getEntryFromID(db, ees0.ID)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if eApax.Lemma.Strn != "tjubba" {
		t.Errorf(fs, "tjubba", eApax.Lemma.Strn)
	}
	if eApax.WordParts != "fin+krog" {
		t.Errorf(fs, "fin+krog", eApax.WordParts)
	}
	if eApax.Language != "gummiapa" {
		t.Errorf(fs, "gummiapa", eApax.Language)
	}
	if got, want := len(eApax.EntryValidations), 0; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	if got, want := len(eApax.Comments), 2; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	// Check that no entries with entryvalidation exist
	noev, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{HasEntryValidation: true})
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if got, want := len(noev), 0; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}

	// Throw in tests of entry comment search for
	// lex.EntryComment{Label: "label1", Source: "secret", Comment: "strålande"}
	rezzx, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{CommentLabelLike: "745648w8"})
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if w, g := 0, len(rezzx); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	rezzx, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{CommentLabelLike: "_abel1"})
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if w, g := 1, len(rezzx); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	rezzx, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{CommentSourceLike: "745648w8"})
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if w, g := 0, len(rezzx); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	rezzx, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{CommentSourceLike: "secr_t"})
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if w, g := 1, len(rezzx); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	rezzx, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{CommentLike: "745648w8"})
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if w, g := 0, len(rezzx); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}
	rezzx, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{CommentLike: "%å%"})
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if w, g := 1, len(rezzx); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	// rezz, err := db.Query("select entry.strn from entry where strn regexp '^a'")
	// if err != nil {
	// 	log.Fatalf("Agh: %v", err)
	// }
	// var strn string
	// for rezz.Next() {
	// 	rezz.Scan(&strn)
	// 	log.Printf(">>> %s", strn)
	// }

	// Add another entry with same str as existing one, to test preferred
	e1b := lex.Entry{Strn: "apa",
		PartOfSpeech:   "XX",
		Morphology:     "NEU UTR",
		WordParts:      "apa",
		Language:       "XYZZ",
		Preferred:      true,
		Transcriptions: []lex.Transcription{t1, t2},
		EntryStatus:    lex.EntryStatus{Name: "old2", Source: "tst"}}

	//time.Sleep(2000 * time.Millisecond)

	_, errxb := mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1b})
	if errxb != nil {
		t.Errorf("Failed to insert entry: %v", errxb)
	}

	// Check that only the new entry has Preferred == 1
	//q := Query{Words: []string{"apa"}, Page: 0, PageLength: 25}

	var entries2 []lex.Entry
	entries2, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	//fmt.Printf("%#v\n", entries2[0])
	//fmt.Printf("%#v\n", entries2[1])
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if got, want := len(entries2), 2; got != want {
		t.Errorf(fs, want, got)
	}

	if entries2[0].PartOfSpeech == "XX" && !entries2[0].Preferred {
		t.Errorf(fs, "true", entries2[0].Preferred)
	}
	if entries2[1].PartOfSpeech == "XX" && !entries2[1].Preferred {
		t.Errorf(fs, "true", entries2[1].Preferred)
	}
	if entries2[0].PartOfSpeech != "XX" && entries2[0].Preferred {
		t.Errorf(fs, "false", entries2[0].Preferred)
	}
	if entries2[1].PartOfSpeech != "XX" && entries2[1].Preferred {
		t.Errorf(fs, "false", entries2[1].Preferred)
	}

	// TODO should be in a test of its own
	eStatsus, err := mariaDBIF{}.listCurrentEntryStatuses(db, l.name)
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := 2, len(eStatsus); w != g {
		t.Errorf(fs, w, g)
	}
	eStatsus2, err := mariaDBIF{}.listAllEntryStatuses(db, l.name)
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := 3, len(eStatsus2); w != g {
		t.Errorf(fs, w, g)
	}

	stat, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{EntryStatus: []string{"new"}})
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := 1, len(stat); w != g {
		t.Errorf(fs, w, g)
	}
	stat1, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{EntryStatus: []string{"dkhfkhekjeh"}})
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := 0, len(stat1); w != g {
		t.Errorf(fs, w, g)
	}
	stat2, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, Query{EntryStatus: []string{"new", "old2"}})
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := 2, len(stat2); w != g {
		t.Errorf(fs, w, g)
	}

}

func TestMariadbUnique(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	in := []int64{1, 2, 3}

	res := unique(in)
	if len(res) != 3 {
		t.Errorf(fs, 3, len(res))
	}

	in = []int64{3, 3, 3}

	res = unique(in)
	if len(res) != 1 {
		t.Errorf(fs, 1, len(res))
	}
	if res[0] != 3 {
		t.Errorf(fs, 3, res[0])
	}
}

func TestMariadbImportLexiconFile(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	symbolSet, err := symbolset.LoadSymbolSet("./test_data/sv-se_ws-sampa.sym")
	if err != nil {
		log.Fatal(err)
	}

	// dbFile := "./iotestlex.db"
	// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
	// 	err := os.Remove(dbFile)
	// 	ff("failed to remove iotestlex.db : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", "./iotestlex.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test2")
	if err != nil {
		log.Fatal(err)
	}

	//defer db.Commit()
	defer db.Close()

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	// defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	ff("Failed to create lexicon db: %v", err)

	logger := StderrLogger{}
	l := lexicon{name: "test", symbolSetName: symbolSet.Name, locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	// actual tests start here
	err = ImportMariaDBLexiconFile(db, lex.LexName(l.name), logger, "./sv-lextest.txt", &validation.Validator{})
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	q := Query{Words: []string{"sprängstoff"}}

	res, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf("lookUpIntoSlice : %v", err)
	}

	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o := res[0].Strn
	if o != "sprängstoff" {
		t.Errorf(fs, "sprängstoff", o)
	}

	q = Query{Words: []string{"sittriktiga"}}
	res, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o = res[0].Strn
	if o != "sittriktiga" {
		t.Errorf(fs, "sittriktiga", o)
	}

	//Let's throw in a test of deleteEntry as well:
	eX := res[0]
	mariaDBIF{}.deleteEntry(db, eX.ID, l.name)

	// Run same query again, efter deleting Entry
	resX, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(resX) != 0 {
		t.Errorf(fs, "0", len(res))
	}

}

func Test_ImportLexiconFileWithDupLines(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	symbolSet, err := symbolset.LoadSymbolSet("./test_data/sv-se_ws-sampa.sym")
	if err != nil {
		log.Fatal(err)
	}

	// dbFile := "./iotestlex.db"
	// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
	// 	err := os.Remove(dbFile)
	// 	ff("failed to remove iotestlex.db : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", "./iotestlex.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test3")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	ff("Failed to create lexicon db: %v", err)

	logger := StderrLogger{}
	l := lexicon{name: "test", symbolSetName: symbolSet.Name, locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	// actual tests start here
	err = ImportMariaDBLexiconFile(db, lex.LexName(l.name), logger, "./sv-lextest-dups.txt", &validation.Validator{})
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	q := Query{Words: []string{"sprängstoff"}}

	res, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf("lookUpIntoSlice : %v", err)
	}

	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o := res[0].Strn
	if o != "sprängstoff" {
		t.Errorf(fs, "sprängstoff", o)
	}

	q = Query{Words: []string{"sittriktiga"}}
	res, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o = res[0].Strn
	if o != "sittriktiga" {
		t.Errorf(fs, "sittriktiga", o)
	}

	q = Query{Words: []string{"vadare"}}
	res, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o = res[0].Strn
	if o != "vadare" {
		t.Errorf(fs, "vadare", o)
	}

	q = Query{WordLike: "%"}
	res, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 19 {
		t.Errorf(fs, "19", len(res))
	}
}

func Test_ImportLexiconFileInvalid(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	symbolSet, err := symbolset.LoadSymbolSet("./test_data/sv-se_ws-sampa.sym")
	if err != nil {
		log.Fatal(err)
	}

	//symbolSet := ssMapper.From

	// dbFile := "./iotestlex.db"
	// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
	// 	err := os.Remove(dbFile)
	// 	ff("failed to remove iotestlex.db : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", "./iotestlex.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test4")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	ff("Failed to create lexicon db: %v", err)

	logger := StderrLogger{}
	l := lexicon{name: "test", symbolSetName: symbolSet.Name, locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	// actual tests start here
	err = ImportMariaDBLexiconFile(db, lex.LexName(l.name), logger, "./sv-lextest-invalid-no-fields.txt", &validation.Validator{})
	if err == nil {
		t.Errorf("Expected errors, but got nil")
	}

}

func Test_ImportLexiconFileGz(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	symbolSet, err := symbolset.LoadSymbolSet("./test_data/sv-se_ws-sampa.sym")
	if err != nil {
		log.Fatal(err)
	}

	// dbFile := "./iotestlex.db"
	// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
	// 	err := os.Remove(dbFile)
	// 	ff("failed to remove iotestlex.db : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", "./iotestlex.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test5")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	ff("Failed to create lexicon db: %v", err)

	logger := StderrLogger{}
	l := lexicon{name: "test", symbolSetName: symbolSet.Name, locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	// actual tests start here
	err = ImportMariaDBLexiconFile(db, lex.LexName(l.name), logger, "./sv-lextest.txt.gz", &validation.Validator{})
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	q := Query{Words: []string{"sprängstoff"}}

	res, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o := res[0].Strn
	if o != "sprängstoff" {
		t.Errorf(fs, "sprängstoff", o)
	}

	q = Query{Words: []string{"sittriktiga"}}
	res, err = mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o = res[0].Strn
	if o != "sittriktiga" {
		t.Errorf(fs, "sittriktiga", o)
	}

	//Let's throw in a test of deleteEntry as well:
	eX := res[0]
	mariaDBIF{}.deleteEntry(db, eX.ID, l.name)

	// Run same query again, efter deleting Entry
	resX, err := mariaDBIF{}.lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(resX) != 0 {
		t.Errorf(fs, "0", len(res))
	}

}

// Test below can be used to load big lexicon
// go test -timeout 60m -v -run Test_ImportLexiconBigFileGz

/*
func Test_ImportLexiconBigFileGz(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}

	symbolSet, err := symbolset.LoadSymbolSet("./test_data/sv-se_ws-sampa.sym")
	if err != nil {
		log.Fatal(err)
	}

	// dbFile := "./iotestlex.db"
	// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
	// 	err := os.Remove(dbFile)
	// 	ff("failed to remove iotestlex.db : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", "./iotestlex.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_sv_nst")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database
	ff("Failed to create lexicon db: %v", err)

	logger := StderrLogger{}
	l := lexicon{name: "test", symbolSetName: symbolSet.Name, locale: "ll"}

	l, err = defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	// actual tests start here
	err = ImportMariaDBLexiconFile(db, lex.LexName(l.name), logger, "/home/nikolaj/gitrepos/wikispeech-lexdata/sv-se/nst/swe030224NST.pron-ws.utf8.gz", &validation.Validator{})
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	q := Query{Words: []string{"sprängstoffet"}}

	res, err := lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o := res[0].Strn
	if o != "sprängstoffet" {
		t.Errorf(fs, "sprängstoffet", o)
	}

	q = Query{Words: []string{"sittriktiga"}}
	res, err = lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	if err != nil {
		t.Errorf(fs, nil, err)
	}
	if len(res) != 1 {
		t.Errorf(fs, "1", len(res))
	}
	o = res[0].Strn
	if o != "sittriktiga" {
		t.Errorf(fs, "sittriktiga", o)
	}

	//Let's throw in a test of deleteEntry as well:
	// eX := res[0]
	// deleteEntry(db, eX.ID, l.name)

	// // Run same query again, efter deleting Entry
	// resX, err := lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
	// if err != nil {
	// 	t.Errorf(fs, nil, err)
	// }
	// if len(resX) != 0 {
	// 	t.Errorf(fs, "0", len(res))
	// }

}
*/

/*
	func Test_ImportLexiconBigFileGzPostTest(t *testing.T) {
		if !*WithMariaDB {
			t.Skip("skipping test for mariadb")
			return
		}

		symbolSet, err := symbolset.LoadSymbolSet("./test_data/sv-se_ws-sampa.sym")
		if err != nil {
			log.Fatal(err)
		}

		// dbFile := "./iotestlex.db"
		// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
		// 	err := os.Remove(dbFile)
		// 	ff("failed to remove iotestlex.db : %v", err)
		// }

		// db, err := sql.Open("sqlite3_with_regexp", "./iotestlex.db")
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// _, err = db.Exec("PRAGMA foreign_keys = ON")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
		// ff("Failed to exec PRAGMA call %v", err)

		db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_sv_nst")
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		// _, err = execSchemaMariadb(db) // Creates new lexicon database
		// ff("Failed to create lexicon db: %v", err)

		// logger := StderrLogger{}
		l := lexicon{name: "test", symbolSetName: symbolSet.Name, locale: "ll"}

		// l, err = defineLexicon(db, l)
		// if err != nil {
		// 	t.Errorf(fs, nil, err)
		// }

		// // actual tests start here
		// err = ImportMariaDBLexiconFile(db, lex.LexName(l.name), logger, "/home/nikolaj/gitrepos/wikispeech-lexdata/sv-se/nst/swe030224NST.pron-ws.utf8.gz", &validation.Validator{})
		// if err != nil {
		// 	t.Errorf(fs, nil, err)
		// }

		q := Query{Words: []string{"sprängstoffet"}}

		res, err := lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
		if err != nil {
			t.Errorf(fs, nil, err)
		}

		if len(res) != 1 {
			t.Errorf(fs, "1", len(res))
		}
		o := res[0].Strn
		if o != "sprängstoffet" {
			t.Errorf(fs, "sprängstoffet", o)
		}

		q = Query{Words: []string{"sittriktig"}}
		res, err = lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
		if err != nil {
			t.Errorf(fs, nil, err)
		}
		if len(res) != 1 {
			t.Errorf(fs, "1", len(res))
		}
		o = res[0].Strn
		if o != "sittriktig" {
			t.Errorf(fs, "sittriktig", o)
		}

		// Doesn't work very well when you want to keep the big db...

		//Let's throw in a test of deleteEntry as well:
		// eX := res[0]
		// deleteEntry(db, eX.ID, l.name)

		// // Run same query again, efter deleting Entry
		// resX, err := lookUpIntoSlice(db, []lex.LexName{lex.LexName(l.name)}, q)
		// if err != nil {
		// 	t.Errorf(fs, nil, err)
		// }
		// if len(resX) != 0 {
		// 	t.Errorf(fs, "0", len(res))
		// }

}
*/
func Test_UpdateComments(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	// dbPath := "./testlex_updatecomments.db"

	// if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
	// 	err := os.Remove(dbPath)
	// 	ff("failed to remove "+dbPath+" : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", dbPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test6")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database

	ff("Failed to create lexicon db: %v", err)

	l := lexicon{name: "test", symbolSetName: "ZZ", locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	// TEST UPDATE COMMENTS
	e1 := lex.Entry{Strn: "apa",
		PartOfSpeech: "NN",
		Morphology:   "NEU UTR",
		WordParts:    "apa",
		Language:     "XYZZ",
		Preferred:    true,
		Transcriptions: []lex.Transcription{
			{Strn: "A: p a", Language: "Svetsko"},
			{Strn: "a pp a", Language: "svinspråket"},
		},
		Comments: []lex.EntryComment{
			{Label: "label1", Source: "anon", Comment: "strålande 1"},
		},
		EntryStatus: lex.EntryStatus{Name: "old1", Source: "tst"}}

	_, err = mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1})
	if err != nil {
		t.Errorf(fs, "nil", err)
	}

	que := Query{WordLike: "apa"}
	var addeds lex.EntrySliceWriter
	err = mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName(l.name)}, que, &addeds)
	if err != nil {
		t.Errorf("Wanted nil, got %v", err)
	}
	if got, want := len(addeds.Entries), 1; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	added := addeds.Entries[0]

	added.Comments = []lex.EntryComment{
		{Label: "label2", Source: "anon", Comment: "strålande 2"},
	}
	newE, updated, err := mariaDBIF{}.updateEntry(db, added)

	if err != nil {
		t.Errorf(fs, "nil", err)
	}
	if !updated {
		t.Errorf(fs, true, updated)
	}
	if want, got := true, newE.Strn == e1.Strn; !got {
		t.Errorf(fs, got, want)
	}

	if len(newE.Comments) != 1 || len(newE.Comments) != len(added.Comments) {
		t.Errorf(fs, newE.Comments, added.Comments)
	} else {
		for i, newC := range newE.Comments {
			c := added.Comments[i]
			if c.Comment != newC.Comment || c.Label != newC.Label || c.Source != newC.Source {
				t.Errorf(fs, newC, c)
			}
		}
	}
}

func Test_ValidationRuleLike(t *testing.T) {
	if !*WithMariaDB {
		t.Skip("skipping test for mariadb")
		return
	}
	// dbPath := "./testlex_validationrulelike.db"

	// if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
	// 	err := os.Remove(dbPath)
	// 	ff("failed to remove "+dbPath+" : %v", err)
	// }

	// db, err := sql.Open("sqlite3_with_regexp", dbPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("PRAGMA foreign_keys = ON")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("PRAGMA case_sensitive_like=ON")
	// ff("Failed to exec PRAGMA call %v", err)

	db, err := sql.Open("mysql", "speechoid:@tcp(127.0.0.1:3306)/wikispeech_pronlex_test7")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = execSchemaMariadb(db) // Creates new lexicon database

	ff("Failed to create lexicon db: %v", err)

	l := lexicon{name: "test", symbolSetName: "ZZ", locale: "ll"}

	l, err = mariaDBIF{}.defineLexicon(db, l)
	if err != nil {
		t.Errorf(fs, nil, err)
	}

	e1 := lex.Entry{Strn: "apa1",
		PartOfSpeech: "NN",
		Morphology:   "NEU UTR",
		WordParts:    "apa",
		Language:     "XYZZ",
		Preferred:    true,
		Transcriptions: []lex.Transcription{
			{Strn: "A: p a", Language: "Svetsko"},
			{Strn: "a pp a", Language: "svinspråket"},
		},
		Comments: []lex.EntryComment{
			{Label: "label1", Source: "anon", Comment: "strålande 1"},
		},
		EntryStatus: lex.EntryStatus{Name: "old1", Source: "tst"},
		EntryValidations: []lex.EntryValidation{
			{RuleName: "rule1", Level: "fatal", Message: "nizze"},
		},
	}

	e2 := lex.Entry{Strn: "apa2",
		PartOfSpeech: "NN",
		Morphology:   "NEU UTR",
		WordParts:    "apa",
		Language:     "XYZZ",
		Preferred:    true,
		Transcriptions: []lex.Transcription{
			{Strn: "A: p a", Language: "Svetsko"},
			{Strn: "a pp a", Language: "svinspråket"},
		},
		Comments: []lex.EntryComment{
			{Label: "label1", Source: "anon", Comment: "strålande 1"},
		},
		EntryStatus: lex.EntryStatus{Name: "old1", Source: "tst"},
		EntryValidations: []lex.EntryValidation{
			{RuleName: "rule2", Level: "fatal", Message: "nizze"},
		},
	}

	_, err = mariaDBIF{}.insertEntries(db, l, []lex.Entry{e1, e2})
	if err != nil {
		t.Errorf(fs, "nil", err)
	}

	que1 := Query{ValidationRuleLike: "rule%"}
	var searchRes1 lex.EntrySliceWriter
	err = mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName(l.name)}, que1, &searchRes1)
	if err != nil {
		t.Errorf("Wanted nil, got %v", err)
	}
	if got, want := len(searchRes1.Entries), 2; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	if !reflect.DeepEqual(searchRes1.Entries[0].Strn, e1.Strn) {
		t.Errorf("Got: %v Wanted: %v", searchRes1, e1)
	}
	if !reflect.DeepEqual(searchRes1.Entries[1].Strn, e2.Strn) {
		t.Errorf("Got: %v Wanted: %v", searchRes1, e2)
	}

	que2 := Query{ValidationRuleLike: "rule1"}
	var searchRes2 lex.EntrySliceWriter
	err = mariaDBIF{}.lookUp(db, []lex.LexName{lex.LexName(l.name)}, que2, &searchRes2)
	if err != nil {
		t.Errorf("Wanted nil, got %v", err)
	}
	if got, want := len(searchRes2.Entries), 1; got != want {
		t.Errorf("Got: %v Wanted: %v", got, want)
	}
	if !reflect.DeepEqual(searchRes2.Entries[0].Strn, e1.Strn) {
		t.Errorf("Got: %v Wanted: %v", searchRes2, e1)
	}
}
