package dbapi

import (
	"bufio"
	"compress/gzip"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/line"
	"github.com/stts-se/pronlex/validation"
)

// ImportSqliteLexiconFile is intended for 'clean' imports. It doesn't check whether the words already exist and so on. It does not do any sanity checks whatsoever of the transcriptions before they are added. If the validator parameter is initialized, each entry will be validated before import, and the validation result will be added to the db.
func ImportSqliteLexiconFile(db *sql.DB, lexiconName lex.LexName, logger Logger, lexiconFileName string, validator *validation.Validator) error {
	return importLexiconFile(sqliteDBIF{}, db, lexiconName, logger, lexiconFileName, validator)
}

// ImportMariDBLexiconFile is intended for 'clean' imports. It doesn't check whether the words already exist and so on. It does not do any sanity checks whatsoever of the transcriptions before they are added. If the validator parameter is initialized, each entry will be validated before import, and the validation result will be added to the db.
func ImportMariaDBLexiconFile(db *sql.DB, lexiconName lex.LexName, logger Logger, lexiconFileName string, validator *validation.Validator) error {
	return importLexiconFile(mariaDBIF{}, db, lexiconName, logger, lexiconFileName, validator)
}

// importLexiconFile is intended for 'clean' imports. It doesn't check whether the words already exist and so on. It does not do any sanity checks whatsoever of the transcriptions before they are added. If the validator parameter is initialized, each entry will be validated before import, and the validation result will be added to the db.
func importLexiconFile(dbif DBIF, db *sql.DB, lexiconName lex.LexName, logger Logger, lexiconFileName string, validator *validation.Validator) error {

	logger.Write(fmt.Sprintf("lexiconName: %v", lexiconName))
	logger.Write(fmt.Sprintf("lexiconFileName: %v", lexiconFileName))
	logger.Write(fmt.Sprintf("dbif name: %v", dbif.name()))

	if _, err := os.Stat(lexiconFileName); os.IsNotExist(err) {
		var msg = fmt.Sprintf("ImportLexiconFile failed to open file : %v", err)
		logger.Write(msg)
		return fmt.Errorf("%v", msg)
	}

	// TODO sanitise lexiconFileName
	fh, err := os.Open(filepath.Clean(lexiconFileName))
	if err != nil {
		var msg = fmt.Sprintf("ImportLexiconFile failed to open file : %v", err)
		logger.Write(msg)
		return fmt.Errorf("%v", msg)
	}
	/* #nosec G307 */
	defer fh.Close()

	var s *bufio.Scanner
	if strings.HasSuffix(lexiconFileName, ".gz") {
		gz, err := gzip.NewReader(fh)
		if err != nil {
			var msg = fmt.Sprintf("ImportLexiconFile failed to open gz reader : %v", err)
			logger.Write(msg)
			return fmt.Errorf("%v", msg)
		}
		s = bufio.NewScanner(gz)
	} else {
		s = bufio.NewScanner(fh)
	}

	wsFmt, err := line.NewWS()
	if err != nil {
		var msg = fmt.Sprintf("ImportLexiconFile failed to instantiate lexicon line parser : %v", err)
		logger.Write(msg)
		return fmt.Errorf("%v", msg)
	}

	lexicon, err := dbif.getLexicon(db, string(lexiconName))
	if err != nil {
		var msg = fmt.Sprintf("ImportLexiconFile failed to get lexicon id for lexicon: %s : %v", lexiconName, err)
		logger.Write(msg)
		return fmt.Errorf("%v", msg)
	}

	msg := fmt.Sprintf("Trying to load file: %s", lexiconFileName)
	logger.Write(msg)

	var nImported = 0
	var nSkippedDups = 0
	var nTotal = 0
	var eBuf []lex.Entry
	var readLines = make(map[string]bool)
	var readEntries = make(map[string]bool)
	for s.Scan() {
		if err := s.Err(); err != nil {
			var msg = fmt.Sprintf("error when reading lines from lexicon file : %v", err)
			logger.Write(msg)
			return fmt.Errorf("%v", msg)
		}
		l := s.Text()

		if strings.HasPrefix(l, "#") {
			continue
		}
		if l == "" {
			continue
		}

		nTotal++
		if _, ok := readLines[l]; ok {
			//var msg = fmt.Sprintf("Skipping duplicate input line : %s", l)
			nSkippedDups++
			//logger.Write(msg)
			continue
		}
		e, err := wsFmt.ParseToEntry(l)
		if err != nil {
			var msg = fmt.Sprintf("couldn't parse line to entry : %v", err)
			logger.Write(msg)
			return fmt.Errorf("%v", msg)
		}
		readLines[l] = true
		eToString, err := wsFmt.Entry2String(e)
		if err != nil {
			var msg = fmt.Sprintf("couldn't convert entry to string : %v", err)
			logger.Write(msg)
			return fmt.Errorf("%v", msg)
		}
		if _, ok := readEntries[eToString]; ok {
			//var msg = fmt.Sprintf("Skipping duplicate input entry : %v", e)
			nSkippedDups++
			//logger.Write(msg)
			continue
		}

		readEntries[eToString] = true

		if validator != nil && validator.IsDefined() {
			validator.ValidateEntry(&e)
		}

		eBuf = append(eBuf, e)
		if nTotal%1000 == 0 {
			_, err = dbif.insertEntries(db, lexicon, eBuf)
			if err != nil {
				var msg = fmt.Sprintf("ImportLexiconFile failed to insert entries : %v", err)
				logger.Write(msg)
				return fmt.Errorf("%v", msg)
			}
			nImported = nImported + len(eBuf)
			msg2 := fmt.Sprintf("ImportLexiconFile: Inserted entries (total lines imported: %d)", nImported)
			logger.Progress(msg2)
			eBuf = make([]lex.Entry, 0)
		}
		if logger.LogInterval() > 0 && nTotal%logger.LogInterval() == 0 {
			msg2 := fmt.Sprintf("ImportLexiconFile: Lines read: %d                         ", nTotal)
			logger.Progress(msg2)
		}
	}
	_, err = dbif.insertEntries(db, lexicon, eBuf) // flushing the buffer
	if err != nil {
		var msg = fmt.Sprintf("ImportLexiconFile failed to insert entries : %v", err)
		logger.Write(msg)
		return fmt.Errorf("%v", msg)
	} // else
	nImported = nImported + len(eBuf)
	msg2 := fmt.Sprintf("ImportLexiconFile: Inserted entries (total lines imported: %d)", nImported)
	logger.Write(msg2)

	logger.Write("Finalizing import ... ")

	// NB. Syntax below is for sqlite, the syntax for mariadb is different
	if dbif.engine() == Sqlite {
		_, err = db.Exec("ANALYZE")
		if err != nil {
			var msg = fmt.Sprintf("failed to exec analyze cmd to db : %v", err)
			logger.Write(msg)
			return fmt.Errorf("%v", msg)
		}
	}

	msg3 := fmt.Sprintf("Lines read:      %d", nTotal)
	logger.Write(msg3)
	msg3 = fmt.Sprintf("Lines imported:  %d", nImported)
	logger.Write(msg3)
	msg3 = fmt.Sprintf("Lines skipped:   %d (duplicates)", nSkippedDups)
	logger.Write(msg3)

	if err := s.Err(); err != nil {
		msg4 := fmt.Sprintf("ImportLexiconFile failed to instantiate lexicon line parser : %v", err)
		logger.Write(msg4)
		return fmt.Errorf("%v", msg4)
	}

	return nil
}

// PrintMode specified the type of output to print (all/valid/invalid)
type PrintMode int

const (
	// PrintAll prints all entries, valid or invalid
	PrintAll PrintMode = iota

	// PrintValid prints valid entries only
	PrintValid

	// PrintInvalid prints invalid entries only
	PrintInvalid
)

// ValidateLexiconFile validates the input file and prints any validation errors to the specified logger.
func ValidateLexiconFile(logger Logger, lexiconFileName string, validator *validation.Validator, printMode PrintMode) error {

	start := time.Now()

	var wg sync.WaitGroup
	log.Printf("lexiconFileName: %v\n", lexiconFileName)

	if _, err := os.Stat(lexiconFileName); os.IsNotExist(err) {
		var msg = fmt.Sprintf("ValidateLexiconFile failed to open file : %v", err)
		log.Println(msg)
		return fmt.Errorf("%v", msg)
	}

	// TODO santise lexiconFileName
	fh, err := os.Open(filepath.Clean(lexiconFileName))
	if err != nil {
		var msg = fmt.Sprintf("ValidateLexiconFile failed to open file : %v", err)
		log.Println(msg)
		return fmt.Errorf("%v", msg)
	}
	/* #nosec G307 */
	defer fh.Close()

	var s *bufio.Scanner
	if strings.HasSuffix(lexiconFileName, ".gz") {
		gz, err := gzip.NewReader(fh)
		if err != nil {
			var msg = fmt.Sprintf("ValidateLexiconFile failed to open gz reader : %v", err)
			log.Println(msg)
			return fmt.Errorf("%v", msg)
		}
		s = bufio.NewScanner(gz)
	} else {
		s = bufio.NewScanner(fh)
	}

	wsFmt, err := line.NewWS()
	if err != nil {
		var msg = fmt.Sprintf("ValidateLexiconFile failed to instantiate lexicon line parser : %v", err)
		log.Println(msg)
		return fmt.Errorf("%v", msg)
	}

	msg := fmt.Sprintf("Trying to load file: %s", lexiconFileName)
	log.Println(msg)

	n := 0
	nPrinted := 0
	nValid := 0
	for s.Scan() {
		if err := s.Err(); err != nil {
			var msg = fmt.Sprintf("error when reading lines from lexicon file : %v", err)
			log.Println(msg)
			return fmt.Errorf("%v", msg)
		}
		l := s.Text()

		if strings.HasPrefix(l, "#") {
			continue
		}
		if l == "" {
			continue
		}

		e, err := wsFmt.ParseToEntry(l)
		if err != nil {
			var msg = fmt.Sprintf("couldn't parse line to entry : %v", err)
			log.Println(msg)
			return fmt.Errorf("%v", msg)
		}

		wg.Add(1)
		go func(ee lex.Entry) {
			defer wg.Done()
			validator.ValidateEntry(&ee)
			isValid := (len(ee.EntryValidations) == 0)
			if isValid {
				nValid = nValid + 1
			}
			if printMode == PrintValid && isValid {
				logger.Write(l)
				nPrinted = nPrinted + 1
			} else if printMode == PrintAll {
				out := []string{l}
				nPrinted = nPrinted + 1
				for _, v := range ee.EntryValidations {
					out = append(out, fmt.Sprintf("#INVALID\t%#v", v.String()))
				}
				logger.Write(strings.Join(out, "\n"))
			} else if printMode == PrintInvalid && !isValid {
				out := []string{l}
				nPrinted = nPrinted + 1
				for _, v := range ee.EntryValidations {
					out = append(out, fmt.Sprintf("#INVALID\t%#v", v.String()))
				}
				logger.Write(strings.Join(out, "\n"))
			}

			n++
			if logger.LogInterval() > 0 && n%logger.LogInterval() == 0 {
				msg2 := fmt.Sprintf("Lines read: %d", n)
				log.Println(msg2)
			}
		}(e)
	}
	wg.Wait()
	log.Printf("Lines read:\t%d", n)
	log.Printf("Lines printed:\t%d", nPrinted)
	log.Printf("Lines valid:\t%d", nValid)

	end := time.Now()
	log.Printf("dbapi/io.go ValidateLexiconFile took %v\n", end.Sub(start))

	if err := s.Err(); err != nil {
		msg = fmt.Sprintf("ValidateLexiconFile failed to instantiate lexicon line parser : %v", err)
		log.Println(msg)
		return fmt.Errorf("%v", msg)
	}

	return nil
}
