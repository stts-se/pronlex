/*
Package dbapi contains code wrapped around an SQL(ite3) DB.
It is used for inserting, updating and retrieving lexical entries from
a pronunciation lexicon database. A lexical entry is represented by
the dbapi.Entry struct, that mirrors entries of the entry database
table, along with associated tables such as transcription and lemma.
*/
package dbapi

//go get github.com/mattn/go-sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	//"regexp"
	//	"path"
	"sort"
	"strconv"
	"strings"
	//"sync"

	// installs sqlite3 driver
	//"github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stts-se/pronlex/lex"
	//"github.com/stts-se/pronlex/validation"
)

// var remem = struct {
// 	sync.Mutex
// 	re map[string]*regexp.Regexp
// }{
// 	re: make(map[string]*regexp.Regexp),
// }

// var regexMem = func(re, s string) (bool, error) {

// 	remem.Lock()
// 	defer remem.Unlock()
// 	if r, ok := remem.re[re]; ok {
// 		return r.MatchString(s), nil
// 	}

// 	r, err := regexp.Compile(re)
// 	if err != nil {
// 		return false, err
// 	}
// 	remem.re[re] = r
// 	return r.MatchString(s), nil
// }

// Sqlite3WithRegex registers an Sqlite3 driver with regexp support. (Unfortunately quite slow regexp matching)
// func Sqlite3WithRegex() {
// 	// regex := func(re, s string) (bool, error) {
// 	// 	//return regexp.MatchString(re, s)
// 	// 	return regexp.MatchString(re, s)
// 	// }
// 	sql.Register("sqlite3_with_regexp",
// 		&sqlite3.SQLiteDriver{
// 			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
// 				return conn.RegisterFunc("regexp", regexMem, true)
// 			},
// 		})
// }

type mariaDBIF struct {
}

func (mdb mariaDBIF) name() string {
	return "mariadb"
}
func (mdb mariaDBIF) engine() DBEngine {
	return MariaDB
}

// getSchemaVersion retrieves the schema version from the database (as defined in schema_mariadb.go on first load)
func (mdb mariaDBIF) getSchemaVersion(db *sql.DB) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("dbapi.GetSchemaVersion : %v", err)
	}
	defer tx.Commit()
	return mdb.getSchemaVersionTx(tx)
}

func (mdb mariaDBIF) getSchemaVersionTx(tx *sql.Tx) (string, error) {
	var res string

	q := "SELECT name FROM SchemaVersion"
	row := tx.QueryRow(q).Scan(&res)
	if row == sql.ErrNoRows {
		var msg = "dbapi.getSchemaVersionTx : couldn't retrive schema version"
		err := tx.Rollback()
		if err != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err)
		}

		log.Println(msg)
		return res, errors.New(msg)
	}

	return res, nil
}

// ListLexicons returns a list of the lexicons defined in the db
// (i.e., Lexicon structs corresponding to the rows of the lexicon
// table).
//
// TODO: Create a DB struct, and move functions of type
// funcName(db *sql.DB, ...) into methods of the new struct.
func (mdb mariaDBIF) listLexicons(db *sql.DB) ([]lexicon, error) {
	var res []lexicon
	sql := "select id, name, symbolsetname, locale from Lexicon"
	rows, err := db.Query(sql)
	if err != nil {
		return res, fmt.Errorf("db query failed : %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		l := lexicon{}
		err = rows.Scan(&l.id, &l.name, &l.symbolSetName, &l.locale)
		if err != nil {
			return res, fmt.Errorf("scanning row failed : %v", err)
		}
		res = append(res, l)
	}
	err = rows.Err()
	sort.Slice(res, func(i, j int) bool { return res[i].name < res[j].name })
	return res, err
}

// GetLexicon returns a Lexicon struct matching a lexicon name in the db.
// Returns error if no such lexicon name in db
func (mdb mariaDBIF) getLexicon(db *sql.DB, name string) (lexicon, error) {
	tx, err := db.Begin()
	if err != nil {
		return lexicon{}, fmt.Errorf("failed to create transaction : %v", err)
	}
	defer tx.Commit()
	return mdb.getLexiconTx(tx, name)
}

func (mdb mariaDBIF) getLexiconMapTx(tx *sql.Tx) (map[string]bool, error) {
	res := make(map[string]bool)

	rows, err := tx.Query("select name from Lexicon")
	if err != nil {
		return res, fmt.Errorf("failed db select on lexicon table : %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		res[name] = true
		if err != nil {
			return res, fmt.Errorf("failed db select on lexicon table : %v", err)
		}
	}
	return res, err

}

func (mdb mariaDBIF) getLexiconTx(tx *sql.Tx, name string) (lexicon, error) {
	res := lexicon{}
	name0 := strings.ToLower(name)
	var err error

	row := tx.QueryRow("select id, name, symbolsetname from Lexicon where name = ? ", name0).Scan(&res.id, &res.name, &res.symbolSetName)
	if row == sql.ErrNoRows {
		return res, fmt.Errorf("couldn't find lexicon '%s'", name)
	}

	return res, err

}

// DeleteLexicon deletes the lexicon name from the lexicon
// table. Notice that it does not remove the associated entries.
// It should be impossible to delete the Lexicon table entry if associated to any entries.
func (mdb mariaDBIF) deleteLexicon(db *sql.DB, lexName string) error {
	log.Printf("deleteLexicon called with lexicon name %s\n", lexName)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()
	return deleteLexiconTx(tx, lexName)
}

// DeleteLexiconTx deletes the lexicon name from the lexicon
// table. Notice that it does not remove the associated entries.
// It should be impossible to delete the Lexicon table entry if associated to any entries.
func deleteLexiconTx(tx *sql.Tx, lexName string) error {
	// does it exist?
	lexExists, err := lexiconExists(tx, lexName)
	if err != nil {
		msg := fmt.Sprintf("dbapi.DeleteLexiconTx : failed to lookup lexicon from name %s : %v", lexName, err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return errors.New(msg)
	}
	if !lexExists {
		return fmt.Errorf("dbapi.DeleteLexiconTx : no lexicon exists with name : %s", lexName)
	}

	n := 0
	err = tx.QueryRow("select count(*) from Entry, Lexicon where Lexicon.name = ? and Entry.lexiconId = Lexicon.id", lexName).Scan(&n)
	// must always return a row, no need to check for empty row
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("the was no lexicon with name %s : %v", lexName, err)
		}
		return err
	}
	if n > 0 {
		return fmt.Errorf("delete all its entries before deleting a lexicon (number of entries: %s)", strconv.Itoa(n))
	}

	_, err = tx.Exec("delete from Lexicon where name = ?", lexName)
	if err != nil {
		msg := fmt.Sprintf("failed to delete lexicon : %v", err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return errors.New(msg)
	}
	return nil
}

func lexiconExists(tx *sql.Tx, lexName string) (bool, error) {

	var id int64
	err := tx.QueryRow("SELECT id FROM Lexicon WHERE name = ?", lexName).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, fmt.Errorf("dbapi.lexiconExists db query failed : %v", err)
	default:
		return true, nil
	}
}

// // SuperDeleteLexicon deletes the lexicon name from the lexicon
// // table and also whipes all associated entries out of existence.
// // TODO Send progress message to client over websocket
// func superDeleteLexicon(db *sql.DB, lexName string) error {
// 	tx, err := db.Begin()
// 	defer tx.Commit()
// 	if err != nil {
// 		return fmt.Errorf("superDeleteLexicon failed to initiate transaction : %v", err)
// 	}
// 	return superDeleteLexiconTx(tx, lexName)
// }

// // SuperDeleteLexiconTx deletes the lexicon name from the lexicon
// // table and also whipes all associated entries out of existence.
// func superDeleteLexiconTx(tx *sql.Tx, lexName string) error {

// 	log.Println("dbapi.superDeleteLexiconTX was called")

// 	// does it exist?
// 	lexExists, err := lexiconExists(tx, lexName)
// 	if err != nil {

// 		msg := fmt.Sprintf("dbapi.SuperDeleteLexiconTx : failed to lookup lexicon from name %s : %v", lexName, err)
// 		err2 := tx.Rollback()
// 		if err2 != nil {
// 			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
// 		}

// 		return errors.New(msg)
// 	}
// 	if !lexExists {
// 		return fmt.Errorf("dbapi.SuperDeleteLexiconTx : no lexicon exists with name : %s", lexName)
// 	}

// 	// delete entries
// 	_, err = tx.Exec("DELETE FROM Entry WHERE lexiconId IN (SELECT id FROM Lexicon WHERE name = ?)", lexName)
// 	if err != nil {
// 		msg := fmt.Sprintf("dbapi.SuperDeleteLexiconTx : failed to delete entries : %v", err)
// 		err2 := tx.Rollback()
// 		if err2 != nil {
// 			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
// 		}
// 		return errors.New(msg)
// 	}
// 	log.Println("dbapi.superDeleteLexiconTX finished deleting from entry set")

// 	// delete lexicon
// 	_, err = tx.Exec("DELETE FROM Lexicon WHERE name = ?", lexName)

// 	if err != nil {
// 		msg := fmt.Sprintf("dbapi.SuperDeleteLexiconTx : failed to delete lexicon : %v", err)
// 		err2 := tx.Rollback()
// 		if err2 != nil {
// 			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
// 		}

// 		return errors.New(msg)
// 	}

// 	log.Println("dbapi.superDeleteLexiconTX finished deleting from lexicon set")

// 	log.Printf("Deleted lexicon named %s\n", lexName)

// 	return nil
// }

// TODO: Check that lexName exists, or report error
// TODO: Check that entryId exists, or report error
func (mdb mariaDBIF) deleteEntry(db *sql.DB, entryID int64, lexName string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		msg := fmt.Sprintf("dbapi.deleteEntry failed to start db transaction : %v", err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return 0, errors.New(msg)
	}
	defer tx.Commit()

	// Check that lexicon exists
	_, err = mdb.getLexiconTx(tx, lexName)
	if err != nil {
		msg := fmt.Sprintf("dbapi.deleteEntry failed to find lexicon '%s' : %v", lexName, err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return 0, errors.New(msg)
	}

	res, err := tx.Exec("DELETE FROM Entry WHERE  id = ? AND lexiconId IN (SELECT id FROM Lexicon WHERE name = ?)", entryID, lexName)
	if err != nil {
		msg := fmt.Sprintf("dbapi.deleteEntry failed to delete entry with id '%d' from lexicon '%s' : %v", entryID, lexName, err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return 0, errors.New(msg)
	}

	i, err := res.RowsAffected()
	if err != nil {
		msg := fmt.Sprintf("dbapi.deleteEntry failed to call RowsAffected after trying to delete entry with id '%d' from lexicon '%s' : %v", entryID, lexName, err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return 0, errors.New(msg)
	}

	// No db error, entry id or lexicon name may be wrong
	if i == 0 {

		return 0, fmt.Errorf("dbapi.deleteEntry failed to delete entry with id '%d' from lexicon '%s'", entryID, lexName)
	}

	return entryID, nil
}

// DefineLexicon saves the name of a new lexicon to the db.
func (mdb mariaDBIF) defineLexicon(db *sql.DB, l lexicon) (lexicon, error) {
	tx, err := db.Begin()
	if err != nil {
		return lexicon{}, fmt.Errorf("failed to get db transaction : %v", err)
	}
	defer tx.Commit()
	res, err := defineLexiconTx(tx, l)
	//tx.Commit()

	return res, err
}

// DefineLexiconTx saves the name of a new lexicon to the db.
func defineLexiconTx(tx *sql.Tx, l lexicon) (lexicon, error) {

	// TODO: downcase the two first characters in l.locale ?

	if strings.TrimSpace(l.locale) == "" {
		msg := fmt.Sprintf("failed to define lexicon with empty locale : %v", l)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return l, errors.New(msg)
	}
	if strings.TrimSpace(l.symbolSetName) == "" {
		msg := fmt.Sprintf("failed to define lexicon with empty symbolSetName : %v", l)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return l, errors.New(msg)
	}

	res, err := tx.Exec("insert into Lexicon (name, symbolsetname, locale) values (?, ?, ?)", strings.ToLower(l.name), l.symbolSetName, l.locale)
	if err != nil {
		msg := fmt.Sprintf("failed to define lexicon : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return l, errors.New(msg)
	}

	id, err := res.LastInsertId()
	if err != nil {
		msg := fmt.Sprintf("failed to get last insert id : %v", err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return l, errors.New(msg)
	}

	//tx.Commit()

	return lexicon{id: id, name: strings.ToLower(l.name), symbolSetName: l.symbolSetName}, err
}

// MoveResult is returned from the MoveNewEntries function.
// TODO Since it only contains a single int64, this struct is probably not needed.
// Only useful if more info is to be returned.
type MoveResult struct {
	N int64
}

// MoveNewEntries moves lexical entries from the lexicon named
// fromLexicon to the lexicon named toLexicon.  The 'newSource' string is
// the name of the new source of the entries to be moved, and 'newStatus' is
// the name of the new status to set on the moved entries.  Currently,
// source and/or status may not be the empty string. TODO: Maybe it
// should be possible to skip source and status values?
//
// Only "new" entries are moved, i.e., entries with lex.Entry.Strn
// values found in fromLexicon but *not* found in toLexicon.  The
// rationale behind this function is to first create a small
// additional lexicon with new entries (the fromLexicon), that can
// later be appended to the master lexicon (the toLexicon).
func (mdb mariaDBIF) moveNewEntries(db *sql.DB, fromLexicon, toLexicon, newSource, newStatus string) (MoveResult, error) {
	if strings.TrimSpace(newSource) == "" {
		msg := "MoveNewEntries called with the empty 'newSource' argument"
		return MoveResult{}, errors.New(msg)
	}
	if strings.TrimSpace(newStatus) == "" {
		msg := "MoveNewEntries called with the empty 'newStatus' argument"
		return MoveResult{}, errors.New(msg)
	}

	tx, err := db.Begin()
	if err != nil {
		return MoveResult{}, fmt.Errorf("failed to get db transaction : %v", err)
	}
	defer tx.Commit()

	return mdb.moveNewEntriesTx(tx, fromLexicon, toLexicon, newSource, newStatus)
}

// moveNewEntriesTx is documented under MoveNewEntries
func (mdb mariaDBIF) moveNewEntriesTx(tx *sql.Tx, fromLexicon, toLexicon, newSource, newStatus string) (MoveResult, error) {
	if strings.TrimSpace(newSource) == "" {
		msg := "moveNewEntriesTx called with the empty 'newSource' argument"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return MoveResult{}, errors.New(msg)
	}
	if strings.TrimSpace(newStatus) == "" {
		msg := "moveNewEntriesTx called with the empty 'newStatus' argument"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return MoveResult{}, errors.New(msg)
	}

	res := MoveResult{}
	var err error
	fromLex, err := mdb.getLexiconTx(tx, fromLexicon)
	if err != nil {
		msg := fmt.Sprintf("couldn't find lexicon %s : %v", fromLexicon, err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return res, errors.New(msg)
	}
	toLex, err := mdb.getLexiconTx(tx, toLexicon)
	if err != nil {
		msg := fmt.Sprintf("couldn't find lexicon %s : %v", toLexicon, err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return res, errors.New(msg)
	}

	const where = `WHERE Entry.id IN (SELECT a.id FROM (select * from Entry) AS a WHERE a.lexiconId = ?
                       AND NOT EXISTS(SELECT ee.strn FROM (select * from Entry) AS ee WHERE ee.lexiconId = ? AND ee.strn = a.strn))`

	insertQuery := `INSERT INTO EntryStatus (name, source, entryId, current) SELECT ?, ?, Entry.id, '1' FROM Entry ` + where

	// updateQuery0 := `UPDATE entrystatus SET current = 1 AND source = ? AND name = ? ` + where + ` AND entrystatus.entryId = entry.id`
	q0Rez, err := tx.Exec(insertQuery, newStatus, newSource, fromLex.id, toLex.id)
	if err != nil {
		msg := fmt.Sprintf("failed to update entrystatus : %v", err)

		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return res, errors.New(msg)
	}

	//_ = q0Rez

	updateQuery := `UPDATE Entry SET lexiconId = ? ` + where

	//log.Printf("Q: %s\n", updateQuery)

	qRez, err := tx.Exec(updateQuery, toLex.id, fromLex.id, toLex.id)
	if err != nil {
		msg := fmt.Sprintf("failed to update lexiconids : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return res, errors.New(msg)
	}

	//TODO Should this result in an error and rollback?
	q0N, _ := q0Rez.RowsAffected()
	qN, _ := qRez.RowsAffected()
	if q0N != qN {
		log.Printf("dbapi.moveNewEntriesTx: UPDATE and INSERT queries affected different number of rows: %v and %v", q0N, qN)
	}

	if n, err := qRez.RowsAffected(); err == nil {
		res.N = n
	}
	return res, err
}

// TODO move to function?
var entrySTMTMDB = "insert into Entry (lexiconId, strn, language, partofspeech, morphology, wordparts, preferred) values (?, ?, ?, ?, ?, ?, ?)"
var transAfterEntrySTMTMDB = "insert into Transcription (entryId, strn, language, sources) values (?, ?, ?, ?)"

var statusSetCurrentFalse = "UPDATE EntryStatus SET current = 0 WHERE EntryStatus.entryId = ?"
var insertStatusMDB = "INSERT INTO EntryStatus (entryId, name, source) values (?, ?, ?)"

// InsertEntries saves a list of Entries and associates them to Lexicon
// TODO: Change second input argument to string (lexicon name) instead of Lexicon struct.
// TODO change input arg to sql.Tx
func (mdb mariaDBIF) insertEntries(db *sql.DB, l lexicon, es []lex.Entry) ([]int64, error) {

	var ids []int64
	// Transaction -->
	tx, err := db.Begin()
	if err != nil {
		return ids, fmt.Errorf("begin transaction failed : %v", err)
	}
	defer tx.Commit()

	stmt1, err := tx.Prepare(entrySTMTMDB)
	if err != nil {
		return ids, fmt.Errorf("failed prepare : %v", err)
	}
	stmt2, err := tx.Prepare(transAfterEntrySTMTMDB)
	if err != nil {
		return ids, fmt.Errorf("failed prepare : %v", err)
	}

	for _, e := range es {
		//log.Printf("dbapi: insert entry: %#v", e)

		if len(e.Transcriptions) == 0 {
			msg := fmt.Sprintf("cannot insert entry without transcriptions: %#v", e)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return ids, errors.New(msg)
		}
		// convert 'Preferred' into DB integer value
		var pref int64
		if e.Preferred {
			pref = 1
		}

		//TODO: Sqlite trigger doesn't work in MaryDB. Must set previous preferred to false manually
		if e.Preferred {
			var setPreferredFalse = "UPDATE Entry SET preferred = 0 WHERE Entry.strn = ?"
			_, err := tx.Exec(setPreferredFalse, e.Strn)
			if err != nil {
				msg := fmt.Sprintf("failed preferred update of previous entries : %v", err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}
				return ids, errors.New(msg)
			}
		}

		res, err := tx.Stmt(stmt1).Exec(
			l.id,
			strings.ToLower(e.Strn),
			e.Language,
			e.PartOfSpeech,
			e.Morphology,
			e.WordParts,
			pref)
		if err != nil {
			msg := fmt.Sprintf("failed exec : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}

			return ids, errors.New(msg)
		}

		id, err := res.LastInsertId()
		if err != nil {
			msg := fmt.Sprintf("failed last insert id : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}

			return ids, errors.New(msg)
		}
		// We want thelex.Entry to have the right id for inserting lemma assocs below
		e.ID = id

		ids = append(ids, id)

		// res.Close()

		for _, t := range e.Transcriptions {
			_, err := tx.Stmt(stmt2).Exec(id, t.Strn, t.Language, t.SourcesString())
			if err != nil {
				msg := fmt.Sprintf("failed exec : %v", err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}

				return ids, errors.New(msg)
			}
		}

		//log.Printf("%v", e)
		if e.Lemma.Strn != "" { // && "" != e.Lemma.Reading {
			lemma, err := mdb.setOrGetLemma(tx, e.Lemma.Strn, e.Lemma.Reading, e.Lemma.Paradigm)
			if err != nil {

				msg := fmt.Sprintf("failed set or get lemma : %v", err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}

				return ids, errors.New(msg)
			}
			err = mdb.associateLemma2Entry(tx, lemma, e)
			if err != nil {
				msg := fmt.Sprintf("failed lemma to entry assoc: %v", err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}

				return ids, errors.New(msg)
			}
		}

		if e.Tag != "" {
			err = mdb.insertEntryTagTx(tx, e.ID, e.Tag, e.Strn)
			if err != nil {
				msg := fmt.Sprintf("failed to insert entry tag '%s' for '%s': %v", e.Tag, e.Strn, err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}

				return ids, errors.New(msg)
			}
		}

		if trm(e.EntryStatus.Name) != "" {

			//var insertStatus = "INSERT INTO entrystatus (entryid, name, source) values (?, ?, ?)"

			// _, err := tx.Exec(statusSetCurrentFalse, e.ID)
			// if err != nil {
			// 	tx.Rollback()
			// 	return ids, fmt.Errorf("updating lex.EntryStatus.Current failed : %v", err)
			// }
			_, err = tx.Exec(insertStatusMDB, e.ID, strings.ToLower(e.EntryStatus.Name), strings.ToLower(e.EntryStatus.Source)) //, e.EntryStatus.Current) // TODO?
			if err != nil {
				msg := fmt.Sprintf("inserting EntryStatus failed : %v", err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}

				return ids, errors.New(msg)
			}
		}

		err = mdb.insertEntryValidations(tx, e, e.EntryValidations)
		if err != nil {
			msg := fmt.Sprintf("inserting EntryValidations failed : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}

			return ids, errors.New(msg)
		}

		err = mdb.insertEntryComments(tx, e.ID, e.Comments)
		if err != nil {

			msg := fmt.Sprintf("inserting EntryComments failed : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}

			return ids, errors.New(msg)
		}

	}

	//tx.Commit()
	// <- transaction

	return ids, err
}

// Trigger version
//var insertEntryTag = "INSERT INTO EntryTag (entryId, tag) values (?, ?)"

// Trigger-less version
var insertEntryTagMDB = "INSERT INTO EntryTag (entryId, tag, wordForm) values (?, ?, ?)"

// TODO: Test that EntryTags work as expected after removal of triggers

//TODO Add tests

// TODO Add db look-up to see if db uniqueness constraints are
// violated, to return more gentle error (or no error) instead of
// failing and rolling back.
func (mdb mariaDBIF) insertEntryTagTx(tx *sql.Tx, entryID int64, tag string, wordForm string) error {

	tag = strings.TrimSpace(strings.ToLower(tag))

	// No tag, silently do nothing:
	if tag == "" {
		return nil
	}

	var eId int64
	var eTag, wordForm2 string
	// Check if it is already there, then silently do nuttin'
	chkResErr := tx.QueryRow("SELECT entryId, tag, wordForm FROM EntryTag WHERE entryId = ?", entryID).Scan(&eId, &eTag, &wordForm2)
	_ = chkResErr
	//sdkjjks :=

	// Entry already has wanted tag, silently accept the fact
	if eId == entryID && eTag == tag && wordForm == wordForm2 {
		return nil
	}

	// Entry had different tag, report error but do nothing
	if eTag != "" && tag != eTag {
		return fmt.Errorf("insertEntryTag: failed to insert tag '%s' because entry already had tag '%s'", tag, eTag)
	}

	// TODO Check that another entry of the same wordform has this tag

	//err == sql.ErrNoRows

	insert, err := tx.Prepare(insertEntryTagMDB)
	if err != nil {
		// Let caller be responsible for rollback
		//tx.Rollback()
		return fmt.Errorf("failed prepare : %v", err)
	}

	_, err = tx.Stmt(insert).Exec(entryID, tag, wordForm)
	if err != nil {
		// TODO Maybe no rollback?
		// Let caller be responsible for rollback
		//tx.Rollback()
		return fmt.Errorf("failed insert entry tag : %v", err)
	}

	return nil
}

// InsertLemma saves a lex.Lemma to the db, but does not associate it with a lex.Entry
// TODO do we need both InsertLemma and SetOrGetLemma?
func (mdb mariaDBIF) insertLemma(tx *sql.Tx, l lex.Lemma) (lex.Lemma, error) {
	sql := "insert into Lemma (strn, reading, paradigm) values (?, ?, ?)"
	res, err := tx.Exec(sql, l.Strn, l.Reading, l.Paradigm)
	if err != nil {
		err = fmt.Errorf("failed insert lemma "+l.Strn+": %v", err)
		return lex.Lemma{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		err = fmt.Errorf("failed last LastInsertId after insert lemma "+l.Strn+": %v", err)
	}
	l.ID = id
	return l, err
}

// SetOrGetLemma saves a new lex.Lemma to the db, or returns a matching already existing one
// TODO do we need both InsertLemma and SetOrGetLemma?
func (mdb mariaDBIF) setOrGetLemma(tx *sql.Tx, strn string, reading string, paradigm string) (lex.Lemma, error) {
	res := lex.Lemma{}

	sqlS := "select id, strn, reading, paradigm from Lemma where strn = ? and reading = ?"
	err := tx.QueryRow(sqlS, strn, reading).Scan(&res.ID, &res.Strn, &res.Reading, &res.Paradigm)
	switch {
	case err == sql.ErrNoRows:
		return mdb.insertLemma(tx, lex.Lemma{Strn: strn, Reading: reading, Paradigm: paradigm})
	case err != nil:
		return res, fmt.Errorf("setOrGetLemma failed querying db : %v", err)
	}

	return res, err
}

// AssociateLemma2Entry adds a lex.Lemma to anlex.Entry via a linking table
func (mdb mariaDBIF) associateLemma2Entry(db *sql.Tx, l lex.Lemma, e lex.Entry) error {
	sql := "insert into Lemma2Entry (lemmaId, entryId) values (?, ?)"
	_, err := db.Exec(sql, l.ID, e.ID)
	if err != nil {
		err = fmt.Errorf("failed to associate lemma "+l.Strn+" and entry "+e.Strn+":%v", err)
	}
	return err
}

// LookUpIds takes a Query struct, searches the lexicon db, and writes the result to a slice of ids
func (mdb mariaDBIF) lookUpIds(db *sql.DB, lexNames []lex.LexName, q Query) ([]int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize transaction : %v", err)
	}
	defer tx.Commit()
	return mdb.lookUpIdsTx(tx, lexNames, q)
}

// LookUpIdsTx takes a Query struct, searches the lexicon db, and returns a slice of ids
func (mdb mariaDBIF) lookUpIdsTx(tx *sql.Tx, lexNames []lex.LexName, q Query) ([]int64, error) {
	var result []int64

	err := mdb.validateInputLexicons(tx, lexNames, q)
	if err != nil {
		return result, err
	}

	sqlStmt := selectEntryIdsSQL(lexNames, q)

	rows, err := tx.Query(sqlStmt.sql, sqlStmt.values...)
	if err != nil {
		// nothing to rollback here, but may have been called from within another transaction
		msg := fmt.Sprintf("%v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return result, errors.New(msg)
	}
	defer rows.Close()

	for rows.Next() {
		var entryID int64
		err = rows.Scan(
			&entryID,
		)
		if err != nil {
			return result, fmt.Errorf("lookUpIdsTx rows.Scan failed : %v", err)
		}

		result = append(result, entryID)
	}
	if rows.Err() != nil {
		// nothing to rollback here, but may have been called from within another transaction
		msg := fmt.Sprintf("%v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return nil, errors.New(msg)
	}

	return result, nil
}

// LookUp takes a Query struct, searches the lexicon db, and writes the result to the
// lex.EntryWriter.
func (mdb mariaDBIF) lookUp(db *sql.DB, lexNames []lex.LexName, q Query, out lex.EntryWriter) error {
	//log.Printf("dbapi lookUp QUWRY %#v\n\n", q)
	if q.Empty() {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to initialize transaction : %v", err)
	}
	defer tx.Commit()
	return mdb.lookUpTx(tx, lexNames, q, out)
}

func (mdb mariaDBIF) validateInputLexicons(tx *sql.Tx, lexNames []lex.LexName, q Query) error {
	if len(lexNames) == 0 && len(q.EntryIDs) == 0 { // if entry id is specified, we can do the search without the lexicon name
		msg := "cannot perform a search without at least one lexicon specified"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return errors.New(msg)
	}

	lexiconMap, err := mdb.getLexiconMapTx(tx)
	if err != nil {
		// nothing to rollback here, but may have been called from within another transaction
		msg := fmt.Sprintf("%v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return errors.New(msg)
	}
	for _, lexName := range lexNames {
		_, ok := lexiconMap[string(lexName)]
		if !ok {
			msg := fmt.Sprintf("no lexicon exists with name: %s", lexName)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}

			return errors.New(msg)
		}
	}
	return nil
}

// LookUpTx takes a Query struct, searches the lexicon db, and writes the result to the
// EntryWriter.
// TODO: rewrite to go through the result set before building the result. That is, save all structs corresponding to rows in the scanning run, then build the result structure (so that no identical values are duplicated: a result set may have several rows of repeated data)
func (mdb mariaDBIF) lookUpTx(tx *sql.Tx, lexNames []lex.LexName, q Query, out lex.EntryWriter) error {

	//if q.Empty() {
	//	return nil
	//}

	//log.Printf("dbapi lookUpTx QUWRY %#v\n\n", q)

	sqlStmt := selectEntriesSQL(lexNames, q)

	// log.Printf("SQL %v\n\n", sqlStmt)
	// log.Printf("VALUES %v\n\n", sqlStmt.values)

	err := mdb.validateInputLexicons(tx, lexNames, q)
	if err != nil {
		return err
	}

	rows, err := tx.Query(sqlStmt.sql, sqlStmt.values...)
	if err != nil {
		// nothing to rollback here, but may have been called from within another transaction
		msg := fmt.Sprintf("%v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return errors.New(msg)
	}
	defer rows.Close()

	var entryID, preferred int64
	var lexiconName, entryStrn, entryLanguage, partOfSpeech, morphology, wordParts string

	var transcriptionID, transcriptionEntryID int64
	var transcriptionStrn, transcriptionLanguage, transcriptionSources string

	// Optional/nullable values

	var lemmaID sql.NullInt64
	var lemmaStrn, lemmaReading, lemmaParadigm, entryTag sql.NullString

	var entryStatusID sql.NullInt64
	var entryStatusName, entryStatusSource sql.NullString
	var entryStatusTimestamp sql.NullString //sql.NullInt64
	var entryStatusCurrent sql.NullBool

	var entryValidationID sql.NullInt64
	var entryValidationLevel, entryValidationName, entryValidationMessage, entryValidationTimestamp sql.NullString

	var entryCommentID sql.NullInt64
	var entryCommentLabel, entryCommentSource, entryCommentComment sql.NullString

	// transcription ids read so far, in order not to add same trans twice
	transIDs := make(map[int64]int)
	// comment ids
	commentIDs := make(map[int64]int)
	// entry validation ids read so far, in order not to add same validation twice
	valiIDs := make(map[int64]int)

	var currE lex.Entry
	var lastE int64
	lastE = -1
	for rows.Next() {
		err2 := rows.Scan(
			&lexiconName,
			&entryID,
			&entryStrn,
			&entryLanguage,
			&partOfSpeech,
			&morphology,
			&wordParts,
			&preferred,

			&transcriptionID,
			&transcriptionEntryID,
			&transcriptionStrn,
			&transcriptionLanguage,
			&transcriptionSources,

			// Optional, from LEFT JOIN

			&lemmaID,
			&lemmaStrn,
			&lemmaReading,
			&lemmaParadigm,

			&entryTag,

			&entryStatusID,
			&entryStatusName,
			&entryStatusSource,
			&entryStatusTimestamp,
			&entryStatusCurrent,

			&entryValidationID,
			&entryValidationLevel,
			&entryValidationName,
			&entryValidationMessage,
			&entryValidationTimestamp,

			&entryCommentID,
			&entryCommentLabel,
			&entryCommentSource,
			&entryCommentComment,
		)

		if err2 != nil {
			return fmt.Errorf("lookUpTx failed scan rows : %v", err2)
		}

		// new entry starts here.
		//
		// all rows with same entryID belongs to the same entry.
		// rows ordered by entryID
		var pref bool // convert 'preferred' value from DB integer value
		if preferred == 1 {
			pref = true
		}
		if lastE != entryID {
			if lastE != -1 {
				err3 := out.Write(currE)
				if err3 != nil {
					return fmt.Errorf("lookUpTx failed to write to lex.EntryWriter : %v", err3)
				}
			}
			currE = lex.Entry{
				LexRef:       lex.NewLexRef("", lexiconName), // DBRef is not set here (will be set by DBManager)
				ID:           entryID,
				Strn:         entryStrn,
				Language:     entryLanguage,
				PartOfSpeech: partOfSpeech,
				Morphology:   morphology,
				WordParts:    wordParts,
				Preferred:    pref,
				Tag:          entryTag.String,
			}

			// max one lemma per entry
			if lemmaStrn.Valid && trm(lemmaStrn.String) != "" {
				l := lex.Lemma{Strn: lemmaStrn.String}
				if lemmaID.Valid {
					l.ID = lemmaID.Int64
				}
				if lemmaReading.Valid {
					l.Reading = lemmaReading.String
				}
				if lemmaParadigm.Valid {
					l.Paradigm = lemmaParadigm.String
				}
				currE.Lemma = l
			}

			// TODO Only add once per lex.Entry as long as single status?
			// TODO probably should be a slice of statuses?
			// TODO now checks for current = true before adding to lex.Entry
			if entryStatusID.Valid && entryStatusName.Valid && trm(entryStatusName.String) != "" {
				es := lex.EntryStatus{ID: entryStatusID.Int64, Name: entryStatusName.String}
				if entryStatusSource.Valid {
					es.Source = entryStatusSource.String
				}
				if entryStatusTimestamp.Valid {
					es.Timestamp = entryStatusTimestamp.String
				}
				if entryStatusCurrent.Valid {
					es.Current = entryStatusCurrent.Bool
				}
				// only update the lex.Entry with status if current = true
				if entryStatusCurrent.Valid && entryStatusCurrent.Bool {
					currE.EntryStatus = es
				}
			}
		}
		// Things that may appear in several rows of a single lex.Entry below:
		// transcriptions ordered by id so they will be added
		// in correct order
		// Only add transcriptions that are !ok, i.e. not added already
		if _, ok := transIDs[transcriptionID]; !ok {
			currT := lex.Transcription{
				ID:       transcriptionID,
				EntryID:  transcriptionEntryID,
				Strn:     transcriptionStrn,
				Language: transcriptionLanguage,
				//Sources:  strings.Split(transcriptionSources, SourceDelimiter),
			}
			// Sources may be empty string in db
			if trm(transcriptionSources) == "" {
				currT.Sources = make([]string, 0)
			} else {
				// strings.Split returns the empty string if input the empty string
				currT.Sources = strings.Split(transcriptionSources, lex.SourceDelimiter)
			}

			currE.Transcriptions = append(currE.Transcriptions, currT)
			transIDs[transcriptionID]++
		}

		if currE.EntryValidations == nil {
			currE.EntryValidations = []lex.EntryValidation{}
		}
		// zero or more lex.EntryValidations
		if entryValidationID.Valid && entryValidationLevel.Valid && entryValidationName.Valid && entryValidationMessage.Valid && entryValidationTimestamp.Valid {
			if _, ok := valiIDs[entryValidationID.Int64]; !ok {
				currV := lex.EntryValidation{
					ID:        entryValidationID.Int64,
					Level:     entryValidationLevel.String,
					RuleName:  entryValidationName.String,
					Message:   entryValidationMessage.String,
					Timestamp: entryValidationTimestamp.String,
				}
				currE.EntryValidations = append(currE.EntryValidations, currV)
				valiIDs[entryValidationID.Int64]++
			}
		}

		if currE.Comments == nil {
			currE.Comments = []lex.EntryComment{}
		}
		// Zero or more lex.EntryComments
		if entryCommentID.Valid && entryCommentLabel.Valid && entryCommentSource.Valid && entryCommentComment.Valid {
			if _, ok := commentIDs[entryCommentID.Int64]; !ok {
				currCmt := lex.EntryComment{
					ID:      entryCommentID.Int64,
					Label:   entryCommentLabel.String,
					Source:  entryCommentSource.String,
					Comment: entryCommentComment.String,
				}
				currE.Comments = append(currE.Comments, currCmt)
				commentIDs[entryCommentID.Int64]++
			}

		}

		lastE = entryID
	}

	// mustn't forget last entry, or lexicon will shrink by one
	// entry for each export/import...
	//	fmt.Fprintf(out, "%v\n", currE)
	// but only print last entry if there were any entries...
	if lastE > -1 {
		err = out.Write(currE)
		if err != nil {
			return fmt.Errorf("failed to write to lex.EntryWriter : %v", err)
		}
	}
	if rows.Err() != nil {
		// nothing to rollback here, but may have been called from within another transaction
		msg := fmt.Sprintf("%v", rows.Err())
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return errors.New(msg)
	}

	return nil
}

// LookUpIntoSlice is a wrapper around LookUp, returning a slice of Entries
func (mdb mariaDBIF) lookUpIntoSlice(db *sql.DB, lexNames []lex.LexName, q Query) ([]lex.Entry, error) {
	var esw lex.EntrySliceWriter
	err := mdb.lookUp(db, lexNames, q, &esw)
	if err != nil {
		return esw.Entries, fmt.Errorf("failed lookup : %v", err)
	}
	return esw.Entries, nil
}

// LookUpIntoMap is a wrapper around LookUp, returning a map where the
// keys are word forms and the values are slices of Entries. (There may be several entries with the same Strn value.)
func (mdb mariaDBIF) lookUpIntoMap(db *sql.DB, lexNames []lex.LexName, q Query) (map[string][]lex.Entry, error) {
	res := make(map[string][]lex.Entry)
	var esw lex.EntrySliceWriter
	err := mdb.lookUp(db, lexNames, q, &esw)
	if err != nil {
		return res, fmt.Errorf("failed lookup : %v", err)
	}
	for _, e := range esw.Entries {
		es := res[e.Strn]
		es = append(es, e)
		res[e.Strn] = es
	}
	return res, err
}

// GetEntryFromID is a wrapper around LookUp and returns the lex.Entry corresponding to the db id
func (mdb mariaDBIF) getEntryFromID(db *sql.DB, id int64) (lex.Entry, error) {
	res := lex.Entry{}
	q := Query{EntryIDs: []int64{id}}
	esw := lex.EntrySliceWriter{}
	err := mdb.lookUp(db, []lex.LexName{}, q, &esw)
	if err != nil {
		return res, fmt.Errorf("LookUp failed : %v", err)
	}

	if len(esw.Entries) == 0 {
		return res, fmt.Errorf("no entry found with id %d", id)
	}
	if len(esw.Entries) > 1 {
		return res, fmt.Errorf("LookUp resulted in more than one entry")
	}
	return esw.Entries[0], nil

}

// UpdateEntry wraps call to UpdateEntryTx with a transaction, and returns the updated entry, fresh from the db
// TODO Consider how to handle inconsistent input entries
// TODO Full name of DB as input param?
func (mdb mariaDBIF) updateEntry(db *sql.DB, e lex.Entry) (res lex.Entry, updated bool, err error) {
	tx, err := db.Begin()
	if err != nil {
		msg := fmt.Sprintf("failed starting transaction for updating entry : %v", err)
		if tx != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
		}
		return res, updated, errors.New(msg)
	}
	defer tx.Commit()

	updated, err = mdb.updateEntryTx(tx, e)
	if err != nil {
		msg := fmt.Sprintf("failed updating entry : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return res, updated, errors.New(msg)
	}
	err = tx.Commit()
	if err != nil {
		return res, updated, fmt.Errorf("updateEntry failed db commit : %v", err)
	}

	res, err = mdb.getEntryFromID(db, e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed getting updated entry : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}

		return res, updated, errors.New(msg)
	}
	return res, updated, err
}

// UpdateEntryTx updates the fields of an lex.Entry that do not match the
// corresponding values in the db
func (mdb mariaDBIF) updateEntryTx(tx *sql.Tx, e lex.Entry) (updated bool, err error) { // TODO return the updated entry?
	// updated == false
	//dbEntryMap := //GetEntriesFromIDsTx(tx, []int64{(e.ID)})
	var esw lex.EntrySliceWriter
	err = mdb.lookUpTx(tx, []lex.LexName{e.LexRef.LexName}, Query{EntryIDs: []int64{e.ID}}, &esw) //entryMapToEntrySlice(dbEntryMap)
	if err != nil {
		return false, fmt.Errorf("updateEntryTx : %v", err)
	}

	dbEntries := esw.Entries
	if len(dbEntries) == 0 {
		return updated, fmt.Errorf("no entry with id '%d'", e.ID)
	}
	if len(dbEntries) > 1 {

		return updated, fmt.Errorf("very bad error, more than one entry with id '%d'", e.ID)
	}

	updated1, err := mdb.updateTranscriptions(tx, e, dbEntries[0])
	if err != nil {
		return updated1, err
	}
	updated2, err := mdb.updateLemma(tx, e, dbEntries[0])
	if err != nil {
		return updated2, err
	}

	updated3, err := mdb.updateWordParts(tx, e, dbEntries[0])
	if err != nil {
		return updated3, err
	}
	updated4, err := mdb.updateLanguage(tx, e, dbEntries[0])
	if err != nil {
		return updated4, err
	}
	updated5, err := mdb.updateEntryStatus(tx, e, dbEntries[0])
	if err != nil {
		return updated5, err
	}

	updated6, err := mdb.updateEntryValidation(tx, e, dbEntries[0])
	if err != nil {
		return updated6, err
	}

	updated7, err := mdb.updatePreferred(tx, e, dbEntries[0])
	if err != nil {
		return updated7, err
	}

	updated8, err := mdb.updateEntryTag(tx, e, dbEntries[0])
	if err != nil {
		return updated8, err
	}

	updated9, err := mdb.updateEntryComments(tx, e, dbEntries[0])
	if err != nil {
		return updated9, err
	}
	updated10, err := mdb.updatePartOfSpeech(tx, e, dbEntries[0])
	if err != nil {
		return updated10, err
	}

	updated11, err := mdb.updateMorphology(tx, e, dbEntries[0])
	if err != nil {
		return updated11, err
	}

	return updated1 || updated2 || updated3 || updated4 || updated5 || updated6 || updated7 || updated8 || updated9 || updated10 || updated11, err
}

func getTIDs(ts []lex.Transcription) []int64 {
	var res []int64
	for _, t := range ts {
		res = append(res, t.ID)
	}
	return res
}

func equal(ts1 []lex.Transcription, ts2 []lex.Transcription) bool {
	if len(ts1) != len(ts2) {
		return false
	}
	for i := range ts1 {
		//if ts1[i] != ts2[i] {
		// TODO? Define Equal(Transcription) on lex.Transcription
		if !reflect.DeepEqual(ts1[i], ts2[i]) {
			return false
		}
	}

	return true
}

func (mdb mariaDBIF) updateLanguage(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	if e.ID != dbE.ID {
		msg := "new and old entries have different ids"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	if e.Language == dbE.Language {
		return false, nil
	}
	_, err := tx.Exec("update Entry set language = ? where Entry.id = ?", e.Language, e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed language update : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	return true, nil
}

func (mdb mariaDBIF) updatePartOfSpeech(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	if e.ID != dbE.ID {
		msg := "new and old entries have different ids"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	if e.PartOfSpeech == dbE.PartOfSpeech {
		return false, nil
	}
	_, err := tx.Exec("update Entry set partofspeech = ? where Entry.id = ?", e.PartOfSpeech, e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed partofspeech update : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	return true, nil
}

func (mdb mariaDBIF) updateMorphology(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	if e.ID != dbE.ID {
		msg := "new and old entries have different ids"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	if e.Morphology == dbE.Morphology {
		return false, nil
	}
	_, err := tx.Exec("update Entry set morphology = ? where Entry.id = ?", e.Morphology, e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed morphology update : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	return true, nil
}

func (mdb mariaDBIF) updateWordParts(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	if e.ID != dbE.ID {
		msg := "new and old entries have different ids"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	if e.WordParts == dbE.WordParts {
		return false, nil
	}
	_, err := tx.Exec("update Entry set wordparts = ? where Entry.id = ?", e.WordParts, e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed worparts update : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	return true, nil
}

func (mdb mariaDBIF) updatePreferred(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	if e.ID != dbE.ID {
		msg := "new and old entries have different ids"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	if e.Preferred == dbE.Preferred {
		return false, nil
	}

	// convert bool into DB integer
	var pref int64
	if e.Preferred {
		pref = 1
	}

	//TODO: Sqlite trigger doesn't work in MaryDB. Must set previous preferred to false manually
	if e.Preferred {
		var setPreferredFalse = "UPDATE Entry SET preferred = 0 WHERE Entry.strn = ?"
		_, err := tx.Exec(setPreferredFalse, e.Strn)
		if err != nil {
			msg := fmt.Sprintf("failed preferred update of previous entries : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}
	}
	_, err := tx.Exec("update Entry set preferred = ? where Entry.id = ?", pref, e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed preferred update : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	return true, nil
}

func (mdb mariaDBIF) updateLemma(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (updated bool, err error) {
	if e.Lemma == dbE.Lemma {
		return false, nil
	}
	// If e.Lemma uninitialized, and different from dbE, then wipe
	// old lemma from db
	if e.Lemma.ID == 0 && e.Lemma.Strn == "" {
		_, err = tx.Exec("delete from Lemma where Lemma.id = ?", dbE.Lemma.ID)
		if err != nil {
			msg := fmt.Sprintf("failed to delete old lemma : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}
	}
	// Only one alternative left, to update old lemma with new values
	_, err = tx.Exec("update Lemma set strn = ?, reading = ?, paradigm = ? where Lemma.id = ?", e.Lemma.Strn, e.Lemma.Reading, e.Lemma.Paradigm, dbE.Lemma.ID)
	if err != nil {
		msg := fmt.Sprintf("failed to update lemma : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	return true, nil
}

// TODO: Test that EntryTags work as expected after removal of triggers
func (mdb mariaDBIF) updateEntryTag(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	//log.Printf("dbapi debug updateEntryTag called")
	if e.ID != dbE.ID {
		msg := "updateEntryTag: new and old entries have different ids"
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}

	newTag := strings.TrimSpace(strings.ToLower(e.Tag))
	oldTag := strings.TrimSpace(strings.ToLower(dbE.Tag))

	// log.Println("dbapi oldTag", oldTag)
	// log.Println("dbapi newTag", newTag)

	// Nothing to do
	if newTag == oldTag {
		return false, nil
	}

	// Delete current tag if new tag is empty
	if newTag == "" { // && oldTag != ""
		_, err := tx.Exec("DELETE FROM EntryTag WHERE entryId = ?", e.ID)
		if err != nil {
			msg := fmt.Sprintf("updateEntryTag failed to delete old tag : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}
		return true, nil
	}

	sqlRes, err := tx.Exec("UPDATE EntryTag SET tag = ?, wordForm = ? WHERE entryId = ?", newTag, e.Strn, e.ID)
	if err != nil {
		msg := fmt.Sprintf("updateEntryTag failed : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	rows, err := sqlRes.RowsAffected()
	if err != nil {
		msg := fmt.Sprintf("updateEntryTag failed (couldn't count rows affected) : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}
	if rows == 0 {
		_, err := tx.Exec("INSERT into EntryTag (tag, entryId, wordForm) values (?, ?, ?)", newTag, e.ID, e.Strn)
		if err != nil {
			msg := fmt.Sprintf("updateEntryTag failed : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}
	}

	var tagged string
	err = tx.QueryRow("SELECT tag FROM EntryTag WHERE entryId = ?", e.ID).Scan(&tagged)
	if err != nil {
		return false, fmt.Errorf("updateEntryTag query failed : %v", err)
	}

	if tagged != newTag {
		msg := fmt.Sprintf("failed to set new entrytag to %s (found %s)", newTag, tagged)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}

	return true, nil
}

func (mdb mariaDBIF) updateEntryComments(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {

	if len(e.Comments) == 0 && len(dbE.Comments) == 0 {
		return false, nil
	}

	// Update comments if different numbers of comments
	if len(e.Comments) != len(dbE.Comments) {
		err := mdb.insertEntryComments(tx, dbE.ID, e.Comments)
		return true, err
	}

	// New code:
	if !reflect.DeepEqual(e.Comments, dbE.Comments) {
		err := mdb.insertEntryComments(tx, dbE.ID, e.Comments)
		return true, err
	}

	// Removed code:
	// for i, cmt := range e.Comments {
	// 	dbCmt := dbE.Comments[i]
	// 	if cmt.Label == dbCmt.Label && cmt.Source == dbCmt.Source && cmt.Comment == dbCmt.Comment {
	// 		err := insertEntryComments(tx, dbE.ID, e.Comments)
	// 		return true, err
	// 	}
	// }

	return false, nil
}

// TODO move to function
var transSTMTMDB = "insert into Transcription (entryId, strn, language, sources) values (?, ?, ?, ?)"

func (mdb mariaDBIF) updateTranscriptions(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (updated bool, err error) {
	if e.ID != dbE.ID {
		return false, fmt.Errorf("update and db entry id differ")
	}

	if len(e.Transcriptions) == 0 {
		return false, fmt.Errorf("cannot update to an empty list of transcriptions")
	}

	// the easy way would be to simply nuke any transcriptions for
	// the entry and substitute for the new ones, but we only want
	// to save new transcriptions if there are changes

	// If the new and old transcriptions differ, remove the old
	// and inser the new ones
	if !equal(e.Transcriptions, dbE.Transcriptions) {
		transIDs := getTIDs(dbE.Transcriptions)
		// TODO move to a function
		_, err := tx.Exec("delete from Transcription where Transcription.id in "+nQs(len(transIDs)), convI(transIDs)...)
		if err != nil {
			msg := fmt.Sprintf("failed transcription delete : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}
		for _, t := range e.Transcriptions {
			_, err := tx.Exec(transSTMTMDB, e.ID, t.Strn, t.Language, t.SourcesString())
			if err != nil {
				msg := fmt.Sprintf("failed transcription update : %v", err)
				err2 := tx.Rollback()
				if err2 != nil {
					msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
				}

				return false, errors.New(msg)
			}
		}
		// different sets of transcription, new ones inserted
		return true, nil
	}
	// Nothing happened
	return false, err
}

// TODO always insert new status, or only when name and source have changed. Or...?
func (mdb mariaDBIF) updateEntryStatus(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (updated bool, err error) {
	if trm(e.EntryStatus.Name) != "" {

		// _, err := tx.Exec(statusSetCurrentFalse, dbE.ID)
		// if err != nil {
		// 	tx.Rollback()
		// 	return false, fmt.Errorf("failed EntryStatus.Current update : %v", err)
		// }

		// TODO: Trigger that in Sqlite nulled previous CURRENT values doesn't work in MariaDB.
		// We do this manually for now
		//var statusSetCurrentFalse = "UPDATE EntryStatus SET current = 0 WHERE EntryStatus.entryId = ?"
		_, err = tx.Exec(statusSetCurrentFalse, e.ID)
		if err != nil {
			msg := fmt.Sprintf("nulling previous EntryStatus.current failed : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}

			return false, errors.New(msg)
		}

		_, err = tx.Exec(insertStatusMDB, dbE.ID, strings.ToLower(e.EntryStatus.Name), strings.ToLower(e.EntryStatus.Source))
		if err != nil {
			msg := fmt.Sprintf("failed EntryStatus update : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}

		return true, nil
	}

	return false, nil
}

type vali struct {
	level string
	name  string
	msg   string
}

// TODO test me
// returns the validations of e1 not found in e2 as fist return arg
// returns the validations found only in e2, to be removed, as second arg
func newValidations(e1 lex.Entry, e2 lex.Entry) ([]lex.EntryValidation, []lex.EntryValidation) {
	var res1 []lex.EntryValidation
	var res2 []lex.EntryValidation
	e1M := make(map[vali]int)
	e2M := make(map[vali]int)
	for _, v := range e1.EntryValidations {
		e1M[vali{level: v.Level, name: v.RuleName, msg: v.Message}]++
	}
	for _, v := range e2.EntryValidations {
		e2M[vali{level: v.Level, name: v.RuleName, msg: v.Message}]++
	}

	for _, v := range e1.EntryValidations {
		if _, ok := e2M[vali{level: v.Level, name: v.RuleName, msg: v.Message}]; !ok {
			res1 = append(res1, v) // only in e1
		}
	}
	for _, v := range e2.EntryValidations {
		if _, ok := e1M[vali{level: v.Level, name: v.RuleName, msg: v.Message}]; !ok {
			res2 = append(res2, v) // only in e2
		}
	}

	return res1, res2
}

var insValiSQLMDB = "INSERT INTO EntryValidation (entryId, level, name, message) values (?, ?, ?, ?)"

func (mdb mariaDBIF) insertEntryValidations(tx *sql.Tx, e lex.Entry, eValis []lex.EntryValidation) error {
	for _, v := range eValis {
		_, err := tx.Exec(insValiSQLMDB, e.ID, strings.ToLower(v.Level), v.RuleName, v.Message)
		if err != nil {
			msg := fmt.Sprintf("failed to insert EntryValidation : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return errors.New(msg)
		}
	}
	return nil
}

func (mdb mariaDBIF) updateValidation(db *sql.DB, entries []lex.Entry) error {
	tx, err := db.Begin()
	if err != nil {
		msg := fmt.Sprintf("failed starting transaction for updating validation : %v", err)
		if tx != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
		}
		return errors.New(msg)
	}
	defer tx.Commit()

	err = mdb.updateValidationTx(tx, entries)
	if err != nil {
		msg := fmt.Sprintf("failed updating validation : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return errors.New(msg)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("updateValidation failed db commit : %v", err)
	}

	return nil
}

func (mdb mariaDBIF) updateValidationTx(tx *sql.Tx, entries []lex.Entry) error {
	for _, e := range entries {
		_, err := mdb.updateEntryValidationForce(tx, e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mdb mariaDBIF) updateEntryValidationForce(tx *sql.Tx, e lex.Entry) (bool, error) {
	_, err := tx.Exec("DELETE FROM EntryValidation WHERE entryId = ?", e.ID)
	if err != nil {
		msg := fmt.Sprintf("failed deleting EntryValidation : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return false, errors.New(msg)
	}

	err = mdb.insertEntryValidations(tx, e, e.EntryValidations)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (mdb mariaDBIF) updateEntryValidation(tx *sql.Tx, e lex.Entry, dbE lex.Entry) (bool, error) {
	newValidations, removeValidations := newValidations(e, dbE)
	if len(newValidations) == 0 && len(removeValidations) == 0 {
		return false, nil
	}

	err := mdb.insertEntryValidations(tx, dbE, newValidations)
	if err != nil {
		return false, err
	}

	for _, v := range removeValidations {
		_, err := tx.Exec("DELETE FROM EntryValidation WHERE id = ?", v.ID)
		if err != nil {
			msg := fmt.Sprintf("failed deleting EntryValidation : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return false, errors.New(msg)
		}
	}

	return true, nil
}

var delEntryCommentsSQLMDB = "DELETE FROM EntryComment WHERE entryId = ?"
var insEntryCommentSQLMDB = "INSERT INTO EntryComment (entryId, label, source, comment) values (?, ?, ?, ?)"

func (mdb mariaDBIF) insertEntryComments(tx *sql.Tx, eID int64, eComments []lex.EntryComment) error {

	// Delete all old comments, before adding the new
	// TODO Handle old comments that are to be kept in a smoother way.
	_, err := tx.Exec(delEntryCommentsSQLMDB, eID)
	if err != nil {
		msg := fmt.Sprintf("failed deleting EntryComments : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return errors.New(msg)
	}

	for _, cmt := range eComments {
		_, err := tx.Exec(insEntryCommentSQLMDB, eID, cmt.Label, cmt.Source, cmt.Comment)
		if err != nil {
			msg := fmt.Sprintf("failed inserting EntryComment : %v", err)
			err2 := tx.Rollback()
			if err2 != nil {
				msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
			}
			return errors.New(msg)
		}

	}

	return nil
}

func unique(ns []int64) []int64 {
	tmpMap := make(map[int64]int)
	var res []int64
	for _, n := range ns {
		if _, ok := tmpMap[n]; !ok {
			res = append(res, n)
			tmpMap[n]++
		}
	}
	return res
}

func (mdb mariaDBIF) entryCount(db *sql.DB, lexiconName string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return -1, fmt.Errorf("dbapi.EntryCount failed opening db transaction : %v", err)
	}
	defer tx.Commit()

	// number of entries in a lexicon
	var entries int64
	err = tx.QueryRow("SELECT COUNT(*) FROM Entry, Lexicon WHERE Entry.lexiconId = Lexicon.id and Lexicon.name = ?", lexiconName).Scan(&entries)
	if err != nil || err == sql.ErrNoRows {
		return -1, fmt.Errorf("dbapi.entryCount failed QueryRow : %v", err)
	}
	return entries, nil
}

func (mdb mariaDBIF) locale(db *sql.DB, lexiconName string) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("dbapi.EntryCount failed opening db transaction : %v", err)
	}
	defer tx.Commit()

	var locale string
	err = tx.QueryRow("SELECT locale FROM Lexicon WHERE Lexicon.name = ?", lexiconName).Scan(&locale)
	if err != nil || err == sql.ErrNoRows {
		return "", fmt.Errorf("dbapi.locale failed QueryRow : %v", err)
	}
	return locale, nil
}

// ListCurrentEntryUsers returns a list of all names EntryUsers marked 'current' (i.e., the most recent status).
func (mdb mariaDBIF) listCurrentEntryUsers(db *sql.DB, lexiconName string) ([]string, error) {
	return mdb.listEntryUsers(db, lexiconName, true)
}

func (mdb mariaDBIF) listCurrentEntryUsersWithFreq(db *sql.DB, lexiconName string) (map[string]int, error) {
	return mdb.listEntryUsersWithFreq(db, lexiconName, true)
}

// ListCurrentEntryStatuses returns a list of all names EntryStatuses marked 'current' (i.e., the most recent status).
func (mdb mariaDBIF) listCurrentEntryStatuses(db *sql.DB, lexiconName string) ([]string, error) {
	return mdb.listEntryStatuses(db, lexiconName, true)
}

func (mdb mariaDBIF) listCurrentEntryStatusesWithFreq(db *sql.DB, lexiconName string) (map[string]int, error) {
	return mdb.listEntryStatusesWithFreq(db, lexiconName, true)
}

// ListAllEntryStatuses returns a list of all names EntryStatuses, also those that are not 'current'  (i.e., the most recent status).
// In other words, this list potentially includes statuses not in use, but that have been used.
func (mdb mariaDBIF) listAllEntryStatuses(db *sql.DB, lexiconName string) ([]string, error) {
	return mdb.listEntryStatuses(db, lexiconName, false)
}

func (mdb mariaDBIF) listEntryStatuses(db *sql.DB, lexiconName string, onlyCurrent bool) ([]string, error) {
	var res []string

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("dbapi.ListCurrentEntryStatuses : %v", err)
	}
	defer tx.Commit()

	// TODO This query seems a bit slow?
	q := "SELECT DISTINCT EntryStatus.name FROM Lexicon, Entry, EntryStatus WHERE Lexicon.name = ? AND Lexicon.id = Entry.lexiconID and Entry.id = EntryStatus.entryId"
	qOnlyCurrent := " AND EntryStatus.current = 1"
	if onlyCurrent {
		q += qOnlyCurrent
	}

	rows, err := tx.Query(q, lexiconName)
	if err != nil {
		msg := fmt.Sprintf("ListCurrentEntryStatuses : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return res, errors.New(msg)
	}
	defer rows.Close()
	for rows.Next() {
		var statusName string
		err = rows.Scan(&statusName)
		if err != nil {
			return res, fmt.Errorf("listEntryStatuses failed db row scan : %v", err)
		}

		res = append(res, statusName)
	}

	err = rows.Err()
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, err
}

func (mdb mariaDBIF) listEntryStatusesWithFreq(db *sql.DB, lexiconName string, onlyCurrent bool) (map[string]int, error) {
	var res = make(map[string]int)

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("dbapi.ListCurrentEntryStatusesWithFreq: %v", err)
	}
	defer tx.Commit()

	// TODO This query seems a bit slow?
	q := "SELECT DISTINCT EntryStatus.name, COUNT(EntryStatus.name) FROM Lexicon, Entry, EntryStatus WHERE Lexicon.name = ? AND Lexicon.id = Entry.lexiconId and Entry.id = EntryStatus.entryId"
	qOnlyCurrent := " AND EntryStatus.current = 1"
	if onlyCurrent {
		q += qOnlyCurrent
	}
	q += " GROUP BY EntryStatus.name"

	rows, err := tx.Query(q, lexiconName)
	if err != nil {
		msg := fmt.Sprintf("ListCurrentEntryStatusesWithFreq : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return res, errors.New(msg)
	}
	defer rows.Close()
	for rows.Next() {
		var statusName string
		var freq int
		err = rows.Scan(&statusName, &freq)
		if err != nil {
			return res, fmt.Errorf("listEntryStatusesWithFreq failed db rows scan : %v", err)
		}
		res[statusName] = freq
	}

	err = rows.Err()
	return res, err
}

func (mdb mariaDBIF) listEntryUsersWithFreq(db *sql.DB, lexiconName string, onlyCurrent bool) (map[string]int, error) {
	var res = make(map[string]int)

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("dbapi.ListCurrentEntryUsersWithFreq : %v", err)
	}
	defer tx.Commit()

	// TODO This query seems a bit slow?
	q := "SELECT DISTINCT EntryStatus.source, COUNT(EntryStatus.source) FROM Lexicon, Entry, EntryStatus WHERE Lexicon.name = ? AND Lexicon.id = Entry.lexiconId and Entry.id = EntryStatus.entryId"
	qOnlyCurrent := " AND EntryStatus.current = 1"
	if onlyCurrent {
		q += qOnlyCurrent
	}
	q += " GROUP BY EntryStatus.source"

	rows, err := tx.Query(q, lexiconName)
	if err != nil {
		msg := fmt.Sprintf("ListCurrentEntryUsersWithFreq : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return res, errors.New(msg)
	}
	defer rows.Close()
	for rows.Next() {
		var userName string
		var freq int
		err = rows.Scan(&userName, &freq)
		if err != nil {
			return res, fmt.Errorf("listEntryUsersWithFreq failed db rows scan : %v", err)
		}

		//res = append(res, userName)
		res[userName] = freq
	}
	err = rows.Err()
	return res, err
}

func (mdb mariaDBIF) listEntryUsers(db *sql.DB, lexiconName string, onlyCurrent bool) ([]string, error) {
	var res []string

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("dbapi.ListCurrentEntryUsers : %v", err)
	}
	defer tx.Commit()

	// TODO This query seems a bit slow?
	q := "SELECT DISTINCT EntryStatus.source FROM Lexicon, Entry, EntryStatus WHERE Lexicon.name = ? AND Lexicon.id = Entry.lexiconId and Entry.id = EntryStatus.entryId"
	qOnlyCurrent := " AND EntryStatus.current = 1"
	if onlyCurrent {
		q += qOnlyCurrent
	}

	rows, err := tx.Query(q, lexiconName)
	if err != nil {
		msg := fmt.Sprintf("ListCurrentEntryUsers : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return res, errors.New(msg)
	}
	defer rows.Close()
	for rows.Next() {
		var userName string
		err = rows.Scan(&userName)
		if err != nil {
			return res, fmt.Errorf("listEntryUsers failed db rows scan : %v", err)
		}

		res = append(res, userName)
	}
	err = rows.Err()
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, err
}

func (mdb mariaDBIF) listCommentLabels(db *sql.DB, lexiconName string) ([]string, error) {
	var res []string

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("dbapi.ListCommentLabels : %v", err)
	}
	defer tx.Commit()

	q := "SELECT DISTINCT EntryComment.label FROM Lexicon, Entry, EntryComment WHERE Lexicon.name = ? AND Lexicon.id = Entry.lexiconId"

	rows, err := tx.Query(q, lexiconName)
	if err != nil {
		msg := fmt.Sprintf("ListCommentLabels : %v", err)
		err2 := tx.Rollback()
		if err2 != nil {
			msg = fmt.Sprintf("%s : rollback failed : %v", msg, err2)
		}
		return res, errors.New(msg)
	}
	defer rows.Close()
	for rows.Next() {
		var label string
		err = rows.Scan(&label)
		if err != nil {
			return res, fmt.Errorf("listCommentLabels failed db rows scan : %v", err)
		}
		res = append(res, label)
	}

	err = rows.Err()
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, err
}

// LexiconStats calls the database a number of times, gathering different numbers, e.g. on how many entries there are in a lexicon.
func (mdb mariaDBIF) lexiconStats(db *sql.DB, lexName string) (LexStats, error) {
	res := LexStats{Lexicon: lexName}

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("dbapi.LexiconStats failed opening db transaction : %v", err)
	}
	defer tx.Commit()

	lex, err := mdb.getLexiconTx(tx, lexName)
	if err != nil {
		return res, fmt.Errorf("dbapi.LexiconStats failed getting lexicon id : %v", err)
	}
	lexiconID := lex.id

	// t1 := time.Now()

	// number of entries in a lexicon
	var entries int64
	err = tx.QueryRow("SELECT COUNT(*) FROM Entry WHERE Entry.lexiconId = ?", lexiconID).Scan(&entries)
	if err != nil || err == sql.ErrNoRows {
		return res, fmt.Errorf("dbapi.LexiconStats failed QueryRow : %v", err)
	}
	res.Entries = entries

	// number of each type of entry status

	// t2 := time.Now()
	// log.Printf("dbapi.LexiconStats TOTAL COUNT TOOK %v\n", t2.Sub(t1))

	rows, err := tx.Query("select EntryStatus.name, count(EntryStatus.name) from Entry, EntryStatus where Entry.lexiconId = ? and Entry.id = EntryStatus.entryId and EntryStatus.current = 1 group by EntryStatus.name", lexiconID)
	if err != nil {
		return res, fmt.Errorf("db query failed : %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var status string
		var freq int64
		err = rows.Scan(&status, &freq)
		if err != nil {
			return res, fmt.Errorf("scanning row failed : %v", err)
		}

		res.StatusFrequencies = append(res.StatusFrequencies, StatusFreq{Status: status, Freq: freq}) //status+"\t"+freq)
	}
	err = rows.Err()
	if err != nil {
		return res, err
	}

	// t3 := time.Now()
	// log.Printf("dbapi.LexiconStats COUNT PER STATUS TOOK %v\n", t3.Sub(t2))

	valStats, err := mdb.validationStatsTx(tx, lexiconID)
	res.ValStats = valStats

	// t4 := time.Now()
	// log.Printf("dbapi.LexiconStats VAL STATS TOOK %v\n", t4.Sub(t3))

	// _ = t1
	// _ = t4
	// log.Printf("dbapi.LexiconStats STATS TOOK %v\n", t4.Sub(t1))

	return res, err

}

func (mdb mariaDBIF) validationStats(db *sql.DB, lexName string) (ValStats, error) {
	tx, err := db.Begin()
	if err != nil {
		return ValStats{}, fmt.Errorf("dbapi.ValidationStats failed opening db transaction : %v", err)
	}
	defer tx.Commit()

	lex, err := mdb.getLexiconTx(tx, lexName)
	if err != nil {
		return ValStats{}, fmt.Errorf("dbapi.LexiconStats failed getting lexicon id : %v", err)
	}
	lexID := lex.id
	return mdb.validationStatsTx(tx, lexID)
}

func (mdb mariaDBIF) validationStatsTx(tx *sql.Tx, lexiconID int64) (ValStats, error) {

	res := ValStats{Rules: make(map[string]int), Levels: make(map[string]int)}

	// number of entries in the lexicon
	err := tx.QueryRow("SELECT COUNT(*) FROM Entry WHERE Entry.lexiconId = ?", lexiconID).Scan(&res.TotalEntries)
	if err != nil || err == sql.ErrNoRows {
		return res, fmt.Errorf("dbapi.ValidationStats failed QueryRow (1) : %v", err)
	}

	res.ValidatedEntries = res.TotalEntries

	// number of invalid entries
	err = tx.QueryRow("SELECT COUNT(DISTINCT EntryValidation.entryId) FROM Entry, EntryValidation WHERE Entry.id = EntryValidation.entryId AND Entry.lexiconId = ?", lexiconID).Scan(&res.InvalidEntries)
	if err != nil || err == sql.ErrNoRows {
		return res, fmt.Errorf("dbapi.ValidationStats failed QueryRow (2) : %v", err)
	}

	// number of validations
	err = tx.QueryRow("SELECT COUNT(DISTINCT EntryValidation.id) FROM Entry, EntryValidation WHERE Entry.id = EntryValidation.entryId AND Entry.lexiconId = ?", lexiconID).Scan(&res.TotalValidations)
	if err != nil || err == sql.ErrNoRows {
		return res, fmt.Errorf("dbapi.ValidationStats failed QueryRow (3) : %v", err)
	}

	levels, err := tx.Query("select EntryValidation.level, count(EntryValidation.level) from Entry, EntryValidation where Entry.lexiconId = ? and Entry.id = EntryValidation.entryId group by EntryValidation.level", lexiconID)
	if err != nil {
		return res, fmt.Errorf("db query failed : %v", err)
	}

	defer levels.Close()
	for levels.Next() {
		var name string
		var count int
		err = levels.Scan(&name, &count)
		if err != nil {
			return res, fmt.Errorf("scanning row failed : %v", err)
		}

		res.Levels[strings.ToLower(name)] = count
	}
	err = levels.Err()
	if err != nil {
		return res, err
	}

	names, err := tx.Query("select EntryValidation.level, EntryValidation.name, count(EntryValidation.name) from Entry, EntryValidation where Entry.lexiconId = ? and Entry.id = EntryValidation.entryId group by EntryValidation.name", lexiconID)
	if err != nil {
		return res, fmt.Errorf("db query failed : %v", err)
	}

	defer names.Close()
	for names.Next() {
		var name string
		var level string
		var count int
		err = names.Scan(&level, &name, &count)
		nameWithLevel := fmt.Sprintf("%s (%s)", strings.ToLower(name), strings.ToLower(level))
		if err != nil {
			return res, fmt.Errorf("scanning row failed : %v", err)
		}

		res.Rules[nameWithLevel] = count
	}
	err = names.Err()
	if err != nil {
		return res, err
	}

	// finally
	return res, err

}

// func (mdb mariaDBIF) isLexiconDB(dbLocation string, dbRef lex.DBRef) (bool, error) {
// 	db, err := sql.Open("mysql", path.Join(dbLocation, string(dbRef)))
// 	if err != nil {
// 		msg := fmt.Sprintf("dbapi_mariadb: failed to open db : %v", err)
// 		if db != nil {
// 			err2 := db.Close()
// 			if err2 != nil {
// 				msg = fmt.Sprintf("%s : failed to close db : %v", msg, err2)
// 			}
// 		}
// 		return false, errors.New(msg)
// 	}

// 	defer db.Close()

// 	rows, err := db.Query("show tables")
// 	defer rows.Close()
// 	if err != nil {
// 		return false, fmt.Errorf("failed to list tables : %v", err)
// 	}
// 	for rows.Next() {
// 		var db string
// 		err = rows.Scan(&db)
// 		if err != nil {
// 			return false, fmt.Errorf("scanning row failed : %v", err)
// 		}
// 		if db == string(dbRef) {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

func (mdb mariaDBIF) isLexiconDatabase(dbTables []string) bool {
	foundSchemaVersionTable := false
	foundLexiconTable := false
	for _, tbl := range dbTables {
		if tbl == "SchemaVersion" {
			foundSchemaVersionTable = true
		} else if tbl == "Lexicon" {
			foundLexiconTable = true
		}
	}
	return foundSchemaVersionTable && foundLexiconTable
}

func (mdb mariaDBIF) listLexiconDatabases(dbLocation string) ([]lex.DBRef, error) {
	log.Print("dbapi_mariadb: loading dbs from location ", dbLocation)

	var res = []lex.DBRef{}

	db, err := sql.Open("mysql", dbLocation+"/")
	if err != nil {
		msg := fmt.Sprintf("dbapi_mariadb: failed to open db : %v", err)
		if db != nil {
			err2 := db.Close()
			if err2 != nil {
				msg = fmt.Sprintf("%s : failed to close db : %v", msg, err2)
			}
		}
		return res, errors.New(msg)
	}

	defer db.Close()

	dbRows, err := db.Query("show databases")
	//lint:ignore SA5001 we cannot resolve the error anyway
	defer dbRows.Close()
	dbsTmp := []string{}
	if err != nil {
		return res, fmt.Errorf("failed to list databases : %v", err)
	}
	for dbRows.Next() {
		var db string
		err = dbRows.Scan(&db)
		if err != nil {
			return res, fmt.Errorf("scanning row failed : %v", err)
		}
		if db != "information_schema" && !strings.HasPrefix(db, "wikispeech_pronlex_test") {
			dbsTmp = append(dbsTmp, db)
		}
	}

	for _, dbName := range dbsTmp {
		tables := []string{}

		_, err = db.Exec(fmt.Sprintf("use %s", dbName))
		if err != nil {
			return res, fmt.Errorf("failed to use database %s : %v", dbName, err)
		}
		tbRows, err := db.Query("show tables")
		//lint:ignore SA5001 we cannot resolve the error anyway
		defer tbRows.Close()
		if err != nil {
			return res, fmt.Errorf("failed to show tables : %v", err)
		}
		for tbRows.Next() {
			var tbl string
			err = tbRows.Scan(&tbl)
			if err != nil {
				return res, fmt.Errorf("scanning row failed : %v", err)
			}
			tables = append(tables, tbl)
		}
		if mdb.isLexiconDatabase(tables) {
			res = append(res, lex.DBRef(dbName))
		}

	}

	log.Printf("dbapi_mariadb: loaded %v db(s)", len(res))
	return res, nil
}

func (mdb mariaDBIF) dropDB(dbLocation string, dbRef lex.DBRef) error {
	dbPath := filepath.Join(dbLocation, string(dbRef))
	db, err := sql.Open("mysql", dbPath)
	if err != nil {
		msg := fmt.Sprintf("dbapi_mariadb: failed to open db : %v", err)
		if db != nil {
			err2 := db.Close()
			if err2 != nil {
				msg = fmt.Sprintf("%s : failed to close db : %v", msg, err2)
			}
		}
		return errors.New(msg)
	}
	defer db.Close()

	_, err = db.Exec(mariaDBDropTableStmt)
	if err != nil {
		return fmt.Errorf("drop table failed : %v", err)
	}
	return nil
}

func (mdb mariaDBIF) openDB(dbLocation string, dbRef lex.DBRef) (*sql.DB, error) {
	dbPath := filepath.Join(dbLocation, string(dbRef))
	db, err := sql.Open("mysql", dbPath)
	if err != nil {
		msg := fmt.Sprintf("dbapi_mariadb: failed to open db : %v", err)
		if db != nil {
			err2 := db.Close()
			if err2 != nil {
				msg = fmt.Sprintf("%s : failed to close db : %v", msg, err2)
			}
		}
		return db, errors.New(msg)
	}
	return db, nil
}

func (mdb mariaDBIF) defineDB(dbLocation string, dbRef lex.DBRef) error {
	dbPath := filepath.Join(dbLocation, string(dbRef))

	db, err := sql.Open("mysql", dbPath)
	if err != nil {
		msg := fmt.Sprintf("dbapi_mariadb: failed to open db : %v", err)
		if db != nil {
			err2 := db.Close()
			if err2 != nil {
				msg = fmt.Sprintf("%s : failed to close db : %v", msg, err2)
			}
		}
		return errors.New(msg)
	}
	defer db.Close()

	db, err = mdb.openDB(dbLocation, dbRef)
	if err != nil {
		return err
	}

	for _, s := range MariaDBSchema {
		_, err := db.Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mdb mariaDBIF) dbExists(dbLocation string, dbRef lex.DBRef) (bool, error) {
	dbName := string(dbRef)
	db, err := sql.Open("mysql", dbLocation+"/")
	if err != nil {
		msg := fmt.Sprintf("dbapi_mariadb: failed to open db : %v", err)
		if db != nil {
			err2 := db.Close()
			if err2 != nil {
				msg = fmt.Sprintf("%s : failed to close db : %v", msg, err2)
			}
		}
		return false, errors.New(msg)
	}

	defer db.Close()

	rows, err := db.Query("show databases")
	//lint:ignore SA5001 we cannot resolve the error anyway
	defer rows.Close()
	if err != nil {
		return false, fmt.Errorf("failed to list databases : %v", err)
	}
	dbFound := false
	for rows.Next() {
		var db string
		err = rows.Scan(&db)
		if err != nil {
			return false, fmt.Errorf("scanning row failed : %v", err)
		}
		if db == dbName {
			dbFound = true
		}
	}
	if !dbFound {
		return false, nil
	}

	// db found, but does it have the right tables?
	tables := []string{}

	_, err = db.Exec(fmt.Sprintf("use %s", dbName))
	if err != nil {
		return false, fmt.Errorf("failed to use database %s : %v", dbName, err)
	}
	tbRows, err := db.Query("show tables")
	//lint:ignore SA5001 we cannot resolve the error anyway
	defer tbRows.Close()
	if err != nil {
		return false, fmt.Errorf("failed to list tables : %v", err)
	}
	for tbRows.Next() {
		var tbl string
		err = tbRows.Scan(&tbl)
		if err != nil {
			return false, fmt.Errorf("scanning row failed : %v", err)
		}
		tables = append(tables, tbl)
	}
	if mdb.isLexiconDatabase(tables) {
		return true, nil
	}
	return false, nil
}
