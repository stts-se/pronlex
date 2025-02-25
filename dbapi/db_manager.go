package dbapi

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/validation"
)

// DBManager is used by external services (i.e., lexserver) to cache sql database instances along with their names
type DBManager struct {
	mutex        *sync.RWMutex
	dbs          map[lex.DBRef]*sql.DB
	dbif         DBIF
	MaxOpenConns int
}

func (dbm DBManager) Engine() DBEngine {
	return dbm.dbif.engine()
}

func (dbm DBManager) Lock() {
	dbm.mutex.Lock()
}
func (dbm DBManager) Unlock() {
	dbm.mutex.Unlock()
}
func (dbm DBManager) RLock() {
	dbm.mutex.RLock()
}

func (dbm DBManager) RUnlock() {
	dbm.mutex.RUnlock()
}

// NewDBManager creates a new DBManager instance with empty cache
func NewDBManager(engine DBEngine) (*DBManager, error) {
	if engine == Sqlite {
		return NewSqliteDBManager(), nil
	} else if engine == MariaDB {
		return NewMariaDBManager(), nil
	} else {
		return &DBManager{}, fmt.Errorf("unknown db engine: %s", engine.String())
	}
}

// NewSqliteDBManager creates a new DBManager instance with empty cache
func NewSqliteDBManager() *DBManager {
	return &DBManager{mutex: &sync.RWMutex{}, dbs: make(map[lex.DBRef]*sql.DB), dbif: sqliteDBIF{}}
}

// NewMariaDBManager creates a new DBManager instance with empty cache
func NewMariaDBManager() *DBManager {
	return &DBManager{mutex: &sync.RWMutex{}, dbs: make(map[lex.DBRef]*sql.DB), dbif: mariaDBIF{}}
}

// CloseDB is used to close the specified database
func (dbm *DBManager) CloseDB(dbRef lex.DBRef) error {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[dbRef]
	if !ok {
		return fmt.Errorf("DBManager.CloseDB: no such db '%s'", dbRef)
	}
	err := db.Close()
	if err != nil {
		return fmt.Errorf("DBManager.CloseDB: couldn't close '%s'", dbRef)
	}
	log.Printf("DBManager.CloseDB: closed db '%s'", dbRef)
	return err
}

// FirstTimePopulateDBCache reads all available dbs into the database cache
func (dbm *DBManager) FirstTimePopulateDBCache(dbLocation string) error {
	var err error // återanvänds för alla fel

	log.Print("db_manager: loading dbs from location ", dbLocation)
	dbs, err := dbm.dbif.listLexiconDatabases(dbLocation)
	if err != nil {
		return fmt.Errorf("couldn't open db file area: %v", err)
	}

	for _, dbRef := range dbs {
		err := dbm.OpenDB(dbLocation, dbRef)
		if err != nil {
			return fmt.Errorf("db_manager: failed to open db : %v", err)
		}
	}

	log.Printf("db_manager: loaded %v db(s)", len(dbs))
	return nil
}

// DefineDB is used to define a new database and add it to the DB manager cache.
// For Sqlite, the database is created, for MariaDB, it has to be created beforehand by an administrator.
// In both cases, all required tables and triggers are added to the database.
func (dbm *DBManager) DefineDB(dbLocation string, dbRef lex.DBRef) error {
	// TODO: Check that the db doesn't exist???
	// if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
	// 	return fmt.Errorf("dbapi_sqlite: db file already exists : %v", err)
	// }

	// err := dbm.OpenDB(dbRef, dbPath)
	// if err != nil {
	// 	msg := fmt.Sprintf("DBManager.DefineDB: failed to open db : %v", err)
	// 	return errors.New(msg)
	// }

	// db, ok := dbm.dbs[dbRef]
	// if !ok {
	// 	return fmt.Errorf("DBManager.DefineDB: no such db '%s'", dbRef)
	// }

	err := dbm.dbif.defineDB(dbLocation, dbRef)
	if err != nil {
		msg := fmt.Sprintf("DBManager.DefineDB: failed to define db : %v", err)
		return errors.New(msg)
	}

	err = dbm.OpenDB(dbLocation, dbRef)
	if err != nil {
		msg := fmt.Sprintf("DBManager.DefineDB: failed to open db : %v", err)
		return errors.New(msg)
	}

	// db, ok := dbm.dbs[dbRef]
	// if !ok {
	// 	return fmt.Errorf("DBManager.DefineDB: no such db '%s'", dbRef)
	// }
	return nil
}

// OpenDB is used to open an existing database and add it to the DB manager cache.
func (dbm *DBManager) OpenDB(dbLocation string, dbRef lex.DBRef) error {
	name := string(dbRef)
	if name == "" {
		return fmt.Errorf("DBManager.OpenDB: illegal argument: name must not be empty")
	}
	if strings.Contains(name, ":") {
		return fmt.Errorf("DBManager.OpenDB: illegal argument: name must not contain ':'")
	}

	dbm.Lock()
	defer dbm.Unlock()

	if _, ok := dbm.dbs[dbRef]; ok {
		return fmt.Errorf("DBManager.OpenDB: db is already loaded: '%s'", name)
	}

	db, err := dbm.dbif.openDB(dbLocation, dbRef)

	if err != nil {
		return fmt.Errorf("DBManager.OpenDB: couldn't open db : %v", err)
	}

	if dbm.MaxOpenConns > 0 {
		db.SetMaxOpenConns(dbm.MaxOpenConns)
	}

	dbm.dbs[dbRef] = db

	return nil
}

// AddDB is used to add a database to the cached map of available databases. It does NOT create the database on disk. To create AND add the database, use DefineDB instead. To open and add an existing db, use OpenDB
func (dbm *DBManager) AddDB(dbRef lex.DBRef, db *sql.DB) error {
	name := string(dbRef)
	if name == "" {
		return fmt.Errorf("DBManager.AddDB: illegal argument: name must not be empty")
	}
	if strings.Contains(name, ":") {
		return fmt.Errorf("DBManager.AddDB: illegal argument: name must not contain ':'")
	}
	if nil == db {
		return fmt.Errorf("DBManager.AddDB: illegal argument: db must not be nil")
	}

	dbm.Lock()
	defer dbm.Unlock()

	if _, ok := dbm.dbs[dbRef]; ok {
		return fmt.Errorf("DBManager.AddDB: db already exists: '%s'", name)
	}

	dbm.dbs[dbRef] = db

	return nil
}

// RemoveDB is used to remove a database from the cached map of available databases. It does NOT remove from the database from disk.
func (dbm *DBManager) RemoveDB(dbRef lex.DBRef) error {
	name := string(dbRef)
	dbm.Lock()
	defer dbm.Unlock()

	if _, ok := dbm.dbs[dbRef]; !ok {
		return fmt.Errorf("DBManager.RemoveDB: no such db '%s'", name)
	}

	delete(dbm.dbs, dbRef)

	return nil
}

// ContainsDB checks if the input database reference exists
func (dbm *DBManager) ContainsDB(dbRef lex.DBRef) bool {
	_, ok := dbm.dbs[dbRef]
	return ok
}

// ListDBNames lists all database names in the cached map of available databases. It does NOT verify what databases are actually existing on disk.
func (dbm *DBManager) ListDBNames() ([]lex.DBRef, error) {
	var res = []lex.DBRef{}

	dbm.RLock()
	defer dbm.RUnlock()

	for k := range dbm.dbs {
		res = append(res, k)
	}

	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, nil
}

// TODO: Maybe this should be removed. Probably, a db should only be
// possible to remove manually as an administrator.

// // SuperDeleteLexicon deletes the lexicon from the associated lexicon
// // database, and also whipes all associated entries out of existence.
// // Returns an error if the lexicon doesn't exist,
// // TODO Send progress message to client over websocket (it takes some time)
// func (dbm *DBManager) SuperDeleteLexicon(lexRef lex.LexRef) error {
// 	dbm.Lock()
// 	defer dbm.Unlock()

// 	db, ok := dbm.dbs[lexRef.DBRef]
// 	if !ok {
// 		return fmt.Errorf("DBManager.SuperDeleteLexicon: no such db '%s'", lexRef.DBRef)
// 	}

// 	err := superDeleteLexicon(db, string(lexRef.LexName))
// 	if err != nil {
// 		return fmt.Errorf("DBManager.SuperDeleteLexicon: couldn't delete '%s'", lexRef)
// 	}

// 	return nil
// }

// DeleteLexicon deletes the lexicon from the associated lexicon
// database. Returns an error if the lexicon doesn't exist,  or if the lexicon is not empty.
func (dbm *DBManager) DeleteLexicon(lexRef lex.LexRef) error {
	dbm.Lock()
	defer dbm.Unlock()

	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return fmt.Errorf("DBManager.DeleteLexicon: no such db '%s'", lexRef.DBRef)
	}

	err := dbm.dbif.deleteLexicon(db, string(lexRef.LexName))
	if err != nil {
		return fmt.Errorf("DBManager.DeleteLexicon: couldn't delete '%s' : %v", lexRef, err)
	}

	return nil
}

// LexiconStats calls the specified database a number of times, gathering different numbers, e.g. on how many entries there are in a lexicon.
func (dbm *DBManager) LexiconStats(lexRef lex.LexRef) (LexStats, error) {
	dbm.Lock()
	defer dbm.Unlock()

	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return LexStats{}, fmt.Errorf("DBManager.LexiconStats: no such db '%s'", lexRef.DBRef)
	}

	stats, err := dbm.dbif.lexiconStats(db, string(lexRef.LexName))
	if err != nil {
		return LexStats{}, fmt.Errorf("DBManager.LexiconStats: couldn't get stats '%s' : %v", lexRef, err)
	}

	return stats, nil
}

// DefineLexicons saves the names of the new lexicons to the db.
func (dbm *DBManager) DefineLexicons(dbRef lex.DBRef, symbolSetName string, locale string, lexes ...lex.LexName) error {

	dbm.RLock()
	defer dbm.RUnlock()

	for _, l := range lexes {
		db, ok := dbm.dbs[dbRef]
		if !ok {
			return fmt.Errorf("DBManager.DefineLexicon: No such db: '%s'", dbRef)
		}
		_, err := dbm.dbif.defineLexicon(db, lexicon{name: string(l), symbolSetName: symbolSetName, locale: locale})
		if err != nil {
			return fmt.Errorf("DBManager.DefineLexicon: failed to add '%s:%s' : %v", dbRef, l, err)
		}
	}

	return nil
}

// DefineLexicon saves the name of a new lexicon to the db.
func (dbm *DBManager) DefineLexicon(lexRef lex.LexRef, symbolSetName string, locale string) error {

	dbm.RLock()
	defer dbm.RUnlock()

	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return fmt.Errorf("DBManager.DefineLexicon: No such db: '%s'", lexRef.DBRef)
	}
	_, err := dbm.dbif.defineLexicon(db, lexicon{name: string(lexRef.LexName), symbolSetName: symbolSetName, locale: locale})
	if err != nil {
		return fmt.Errorf("DBManager.DefineLexicon: failed to add '%s' : %v", lexRef.String(), err)
	}

	return nil
}

type lookUpRes struct {
	dbRef   lex.DBRef // TODO: move to lex.Entry!!
	entries []lex.Entry
	err     error
}

// ListIDs is a wrapper around lookUpIds, returning a slice of ID's
func (dbm *DBManager) ListIDs(lexRef lex.LexRef) ([]int64, error) {
	dbm.RLock()
	defer dbm.RUnlock()

	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return []int64{}, fmt.Errorf("DBManager.ListIDs failed: no db of name '%s'", lexRef.DBRef)
	}

	ids, err := dbm.dbif.lookUpIds(db, []lex.LexName{lexRef.LexName}, Query{})
	if err != nil {
		return []int64{}, fmt.Errorf("DBManager.ListIDs failed for lexicon : '%s'", lexRef)
	}
	return ids, nil
}

// LookUpIntoSlice is a wrapper around LookUp, returning a slice of Entries
func (dbm *DBManager) LookUpIntoSlice(q DBMQuery) ([]lex.Entry, error) {
	var res = []lex.Entry{}
	writer := lex.EntrySliceWriter{}
	err := dbm.LookUp(q, &writer)
	if err != nil {
		return res, err
	}

	res = append(res, writer.Entries...)

	return res, nil
}

// LookUpIntoMap is a wrapper around LookUp, returning a map of Entries
func (dbm *DBManager) LookUpIntoMap(q DBMQuery) (map[lex.DBRef][]lex.Entry, error) {
	var res = make(map[lex.DBRef][]lex.Entry)
	writer := lex.EntrySliceWriter{}
	err := dbm.LookUp(q, &writer)
	if err != nil {
		return res, err
	}
	for _, e := range writer.Entries {
		es := res[e.LexRef.DBRef]
		es = append(es, e)
		res[e.LexRef.DBRef] = es
	}
	return res, nil
}

// LookUp takes a DBMQuery, searches the specified lexicon for the included search query. The result is written to a lex.EntryWriter.
func (dbm *DBManager) LookUp(q DBMQuery, out lex.EntryWriter) error {
	if len(q.LexRefs) == 0 { //  && len(q.Query.EntryIDs) == 0 {
		return fmt.Errorf("DBManager.LookUp cannot perform a search without at least one lexicon specified (using the 'lexicons' parameter)")
	}

	dbz := make(map[lex.DBRef][]lex.LexName)
	for _, l := range q.LexRefs {
		lexList := dbz[l.DBRef]
		dbz[l.DBRef] = append(lexList, l.LexName)
	}

	dbm.RLock()
	defer dbm.RUnlock()

	ch := make(chan lookUpRes)
	for dbR, lexs := range dbz {
		db, ok := dbm.dbs[dbR]
		if !ok {
			return fmt.Errorf("DBManager.LookUp failed: no db of name '%s'", dbR)
		}

		go func(db0 *sql.DB, dbRef lex.DBRef, lexNames []lex.LexName) {
			rez := lookUpRes{}
			rez.dbRef = dbRef
			ew := lex.EntrySliceWriter{}
			err := dbm.dbif.lookUp(db0, lexNames, q.Query, &ew)
			if err != nil {
				rez.err = fmt.Errorf("dbapi.LookUp failed for %v:%v : %v", dbRef, lexNames, err)
				ch <- rez
				return
			}
			for _, e := range ew.Entries {
				e.LexRef.DBRef = dbRef
				rez.entries = append(rez.entries, e)
			}

			ch <- rez
		}(db, dbR, lexs)
	}

	for i := 0; i < len(dbz); i++ {
		lkUp := <-ch
		if lkUp.err != nil {
			return fmt.Errorf("DBManager.LookUp failed : %v", lkUp.err)
		}

		for _, e := range lkUp.entries {
			err := out.Write(e)
			if err != nil {
				return fmt.Errorf("error writing to lex.EntryWriter : %v", err)
			}
		}
	}

	return nil
}

type lexRes struct {
	lexes []lex.LexRefWithInfo
	err   error
}

// Warning: this is maybe my first attempt at concurrency using a channel in Go

// ListLexicons returns a list of defined lexicons, including database name, lexicon name, and symbol set name
func (dbm *DBManager) ListLexicons() ([]lex.LexRefWithInfo, error) {
	var res = []lex.LexRefWithInfo{}

	dbm.RLock()
	defer dbm.RUnlock()

	ch := make(chan lexRes)
	// defer close(ch) // ?
	dbs := dbm.dbs
	// Go ask each db instance in its own Go-routine
	for dbRef, db := range dbs {
		go func(dbRef lex.DBRef, db *sql.DB, ch0 chan lexRes) {
			lexs, err := dbm.dbif.listLexicons(db)
			lexList := []lex.LexRefWithInfo{}
			for _, ln := range lexs {
				lexRef := lex.LexRef{DBRef: dbRef, LexName: lex.LexName(ln.name)}
				withInfo := lex.LexRefWithInfo{LexRef: lexRef, SymbolSetName: ln.symbolSetName}
				lexList = append(lexList, withInfo)
			}
			r := lexRes{lexes: lexList, err: err}
			ch0 <- r
		}(dbRef, db, ch)
	}

	// Read result from channel
	for i := 0; i < len(dbs); i++ {
		var r lexRes = <-ch // Blocks until there is a result (I
		// think). Can we be stuck here forever, if
		// db call hangs?

		// If we encounter an error, just bail out
		if r.err != nil {
			return res, fmt.Errorf("DBManager.ListLexicons freak out : %v", r.err)

		}

		// A full lexicon name consists of the name of
		// the db and the name of the lexicon in the
		// db joined by ':'
		res = append(res, r.lexes...)

	}

	sort.Slice(res, func(i, j int) bool { return res[i].LexRef.String() < res[j].LexRef.String() })
	close(ch) // ?
	return res, nil
}

// LexiconExists is used to check if the specified lexicon exists in the specified database
func (dbm *DBManager) LexiconExists(lexRef lex.LexRef) (bool, error) {
	lexInfo, err := dbm.ListLexicons()
	if err != nil {
		return false, fmt.Errorf("DBManager.LexiconExists failed : %v", err)
	}
	for _, lf := range lexInfo {
		if lf.LexRef == lexRef {
			return true, nil
		}
	}
	return false, nil
}

// InsertEntries saves a list of Entries and associates them to the lexicon
func (dbm *DBManager) InsertEntries(lexRef lex.LexRef, entries []lex.Entry) ([]int64, error) {

	var res []int64

	dbm.Lock()
	defer dbm.Unlock()

	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return res, fmt.Errorf("DBManager.InsertEntries: unknown db '%s'", lexRef.DBRef)
	}

	//_ = db
	//_ = lexName
	l, err := dbm.dbif.getLexicon(db, string(lexRef.LexName))
	//fmt.Printf("%v\n", l)
	if err != nil {
		return res, fmt.Errorf("DBManager.InsertEntries failed call to getLexicons : %v", err)
	}
	//fmt.Println(lexName)
	res, err = dbm.dbif.insertEntries(db, l, entries)
	if err != nil {
		return res, fmt.Errorf("DBManager.InsertEntries failed: %v", err)
	}
	return res, err
}

// UpdateValidation using the cached validation in the specified lex.Entry
func (dbm *DBManager) UpdateValidation(e lex.Entry) error {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[e.LexRef.DBRef]
	if !ok {
		return fmt.Errorf("DBManager.UpdateValidation: no such db '%s'", e.LexRef.DBRef)
	}

	return dbm.dbif.updateValidation(db, []lex.Entry{e})
}

// UpdateEntry wraps call to UpdateEntryTx with a transaction, and returns the updated entry, fresh from the db
func (dbm *DBManager) UpdateEntry(e lex.Entry) (lex.Entry, bool, error) {
	var res lex.Entry

	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[e.LexRef.DBRef]
	if !ok {
		return res, false, fmt.Errorf("DBManager.UpdateEntry: no such db '%s'", e.LexRef.DBRef)
	}

	return dbm.dbif.updateEntry(db, e)
}

// DeleteEntry deletes an entry from the database
func (dbm *DBManager) DeleteEntry(entryID int64, lexRef lex.LexRef) (int64, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return 0, fmt.Errorf("DBManager.DeleteEntry: no such db '%s'", lexRef.DBRef)
	}

	return dbm.dbif.deleteEntry(db, entryID, string(lexRef.LexName))
}

// ImportLexiconFile is intended for 'clean' imports. It doesn't check whether the words already exist and so on. It does not do any sanity checks whatsoever of the transcriptions before they are added. If the validator parameter is initialized, each entry will be validated before import, and the validation result will be added to the db.
func (dbm *DBManager) ImportLexiconFile(lexRef lex.LexRef, logger Logger, lexiconFileName string, validator *validation.Validator) error {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return fmt.Errorf("DBManager.ImportLexiconFile: no such db '%s'", lexRef.DBRef)
	}
	return importLexiconFile(dbm.dbif, db, lexRef.LexName, logger, lexiconFileName, validator)
}

// EntryCount counts the number of entries in a lexicon
func (dbm *DBManager) EntryCount(lexRef lex.LexRef) (int64, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return 0, fmt.Errorf("DBManager.ImportLexiconFile: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.entryCount(db, string(lexRef.LexName))
}

// Locale looks up the locale for a specific lexicon
func (dbm *DBManager) Locale(lexRef lex.LexRef) (string, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return "", fmt.Errorf("DBManager.ImportLexiconFile: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.locale(db, string(lexRef.LexName))
}

// ListCommentLabels returns a list of all comment labels
func (dbm *DBManager) ListCommentLabels(lexRef lex.LexRef) ([]string, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return []string{}, fmt.Errorf("DBManager.ListCommentLabels: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.listCommentLabels(db, string(lexRef.LexName))
}

// ListCurrentEntryUsers returns a list of all names EntryUsers marked 'current' (i.e., the most recent status).
func (dbm *DBManager) ListCurrentEntryUsers(lexRef lex.LexRef) ([]string, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return []string{}, fmt.Errorf("DBManager.ListCurrentEntryUsers: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.listCurrentEntryUsers(db, string(lexRef.LexName))
}

// ListCurrentEntryUsersWithFreq returns a map of all names EntryUsers marked 'current' (i.e., the most recent status), and the frequency for each user
func (dbm *DBManager) ListCurrentEntryUsersWithFreq(lexRef lex.LexRef) (map[string]int, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return make(map[string]int), fmt.Errorf("DBManager.ListCurrentEntryUsersWithFreq: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.listCurrentEntryUsersWithFreq(db, string(lexRef.LexName))
}

// ListCurrentEntryStatuses returns a list of all names EntryStatuses marked 'current' (i.e., the most recent status).
func (dbm *DBManager) ListCurrentEntryStatuses(lexRef lex.LexRef) ([]string, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return []string{}, fmt.Errorf("DBManager.ListCurrentEntryStatuses: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.listCurrentEntryStatuses(db, string(lexRef.LexName))
}

// ListCurrentEntryStatusesWithFreq returns a list of all names EntryStatuses marked 'current' (i.e., the most recent status), and the frequency for each status.
func (dbm *DBManager) ListCurrentEntryStatusesWithFreq(lexRef lex.LexRef) (map[string]int, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return make(map[string]int), fmt.Errorf("DBManager.ListCurrentEntryStatusesWithFreq: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.listCurrentEntryStatusesWithFreq(db, string(lexRef.LexName))
}

// ListAllEntryStatuses returns a list of all names EntryStatuses, also those that are not 'current'  (i.e., the most recent status).
// In other words, this list potentially includes statuses not in use, but that have been used before.
func (dbm *DBManager) ListAllEntryStatuses(lexRef lex.LexRef) ([]string, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return []string{}, fmt.Errorf("DBManager.ListAllEntryStatuses: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.listAllEntryStatuses(db, string(lexRef.LexName))
}

// GetLexicon returns a information (LexRefWithInfo) matching a lexicon name in the db.
// Returns error if no such lexicon name in db
func (dbm *DBManager) GetLexicon(lexRef lex.LexRef) (lex.LexRefWithInfo, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return lex.LexRefWithInfo{}, fmt.Errorf("DBManager.GetLexicon: no such db '%s'", lexRef.DBRef)
	}
	l, err := dbm.dbif.getLexicon(db, string(lexRef.LexName))
	if err != nil {
		return lex.LexRefWithInfo{}, err
	}
	return lex.LexRefWithInfo{
		LexRef:        lexRef,
		SymbolSetName: l.symbolSetName,
	}, nil
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
func (dbm *DBManager) MoveNewEntries(dbRef lex.DBRef, fromLex, toLex lex.LexName, newSource, newStatus string) (MoveResult, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[dbRef]
	if !ok {
		return MoveResult{}, fmt.Errorf("DBManager.MoveNewEntries: no such db '%s'", dbRef)
	}
	return dbm.dbif.moveNewEntries(db, string(fromLex), string(toLex), newSource, newStatus)
}

// Validate all entries given the specified lexRef and search query. Updates validation stats in db, and returns these.
func (dbm *DBManager) Validate(lexRef lex.LexRef, logger Logger, vd validation.Validator, q Query) (ValStats, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return ValStats{}, fmt.Errorf("DBManager.Validate: no such db '%s'", lexRef.DBRef)
	}
	return validate(dbm.dbif, db, []lex.LexName{lexRef.LexName}, logger, vd, q)
}

// ValidationStats returns existing validation stats for the specified lexRef
func (dbm *DBManager) ValidationStats(lexRef lex.LexRef) (ValStats, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[lexRef.DBRef]
	if !ok {
		return ValStats{}, fmt.Errorf("DBManager.ValidationStats: no such db '%s'", lexRef.DBRef)
	}
	return dbm.dbif.validationStats(db, string(lexRef.LexName))
}

// GetSchemaVersion retrieves the schema version from the database
func (dbm *DBManager) GetSchemaVersion(dbRef lex.DBRef) (string, error) {
	dbm.Lock()
	defer dbm.Unlock()
	db, ok := dbm.dbs[dbRef]
	if !ok {
		return "", fmt.Errorf("DBManager.GetSchemaVersion: no such db '%s'", dbRef)
	}
	return dbm.dbif.getSchemaVersion(db)

}

// DropDB drop the database (cannot be undone).
// For Sqlite, the database is entirely dropped, for MariaDB, all database tables are dropped, but the database is not deleted. Deletion of MariaDB databases should be done by a server admiinstrator.
func (dbm *DBManager) DropDB(dbLocation string, dbRef lex.DBRef) error {
	return dbm.dbif.dropDB(dbLocation, dbRef)
}

// DBExists checks if a database exist. For Sqlite, it checks if the actual database file exists. For MariaDB, it checks if the database exists, and contains tables required for a lexicon database. The reason for this is how the user privileges work for MariaDB. See also DefinedDB and DropDB.
func (dbm *DBManager) DBExists(dbLocation string, dbRef lex.DBRef) (bool, error) {
	return dbm.dbif.dbExists(dbLocation, dbRef)
}
