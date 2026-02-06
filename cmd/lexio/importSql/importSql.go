package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stts-se/pronlex/dbapi"
	"github.com/stts-se/pronlex/lex"
)

// SQLITE
// - DUMP: sqlite3 <dbFile> .dump | gzip -c > <sqlDumpFile>
// - LOAD:  gunzip -c <dumpFile> | sqlite3 <dbFile>

// MARIADB
// - DUMP:  mysqldump -u speechoid -h <dbHost> <dbName> |gzip -c > <sqlDumpFile>
// - LOAD:  gunzip -c <dumpFile> | mysql -u speechoid -h <dbHost> <dbName>

const sqlitePath = "sqlite3"
const mariaDBPath = "mysql"

/*
	func sqlDump(dbFile string, outFile string) error {
		sqliteCmd := exec.Command(sqlitePath, dbFile, ".dump")
		out, err := os.Create(outFile)
		if err != nil {
			log.Fatalf("couldn't create output file %s : %v", outFile, err)
		}
		defer out.Close()
		sqliteCmd.Stdout = out
		sqliteCmd.Stderr = os.Stderr
		err = sqliteCmd.Run()
		if err != nil {
			return err
		}
		log.Printf("Exported %s from db %s\n", dbFile, outFile)
		return nil
	}
*/
var dotAndAfter = regexp.MustCompile("[.].*$")

func validateSchemaVersion(dbm *dbapi.DBManager, dbRef lex.DBRef) error {
	dbVer, err := dbm.GetSchemaVersion(dbRef)
	apiVer := dbapi.SchemaVersion
	if err != nil {
		log.Fatalf("Couldn't retrive schema version : %v", err)
	}
	if dbVer != apiVer {
		dbVerX := dotAndAfter.ReplaceAllString(dbVer, "")
		apiVerX := dotAndAfter.ReplaceAllString(apiVer, "")
		if dbVerX != apiVerX {
			log.Fatalf("Mismatching schema versions. Input file: %s, dbapi.Schema: %s ", dbVer, apiVer)
		} else {
			log.Printf("Compatible schema versions. Input file: %s, dbapi.Schema: %s ", dbVer, apiVer)
		}
	} else {
		log.Printf("Schema version matching dbapi.SchemaVersion: %s\n", apiVer)
	}
	return nil
}

func runPostTests(dbm *dbapi.DBManager, dbLocation string, dbRef lex.DBRef, sqlDumpFile string) {

	//err = dbm.DefineDB(dbm, dbLocation, dbRef)
	err := dbm.OpenDB(dbLocation, dbRef)
	if err != nil {
		log.Fatalf("Couldn't open db: %v", err)
	}
	//lexes, err := dbm.ListLexicons()
	//if err != nil {
	//	log.Fatalf("Failed to list lexicons : %v\n", err)
	//}

	// (1) check schema version
	err = validateSchemaVersion(dbm, dbRef)
	if err != nil {
		log.Fatalf("Couldn't read validate schema version in file %s : %v\n", sqlDumpFile, err)
	}

	// (2) output statistics
	lexes, err := dbm.ListLexicons()
	if err != nil {
		log.Fatalf("Couldn't list lexicons : %v", err)
	}
	for _, lex := range lexes {
		stats, err := dbm.LexiconStats(lex.LexRef)
		if err != nil {
			log.Fatalf("Failed to retrieve statistics : %v", err)
		}
		err = printStats(stats, true)
		if err != nil {
			log.Fatalf("Failed to print statistics : %v", err)
		}
	}

}

func getFileReader(fName string) (io.Reader, error) {
	fs, err := os.Open(filepath.Clean(fName))
	if err != nil {
		log.Fatalf("Couldn't open file %s for reading : %v\n", fName, err)
	}

	if strings.HasSuffix(fName, ".sql") {
		return io.Reader(fs), nil
	} else if strings.HasSuffix(fName, ".sql.gz") {
		if strings.HasSuffix(fName, ".gz") {
			gz, err := gzip.NewReader(fs)
			if err != nil {
				return nil, fmt.Errorf("couldn't to open gz reader : %v", err)
			}
			return io.Reader(gz), nil
		}

	}
	return nil, fmt.Errorf("unknown file type: %s. Expected .sql or .sql.gz", fName)
}

type dsn struct {
	host     string
	user     string
	port     string
	protocol string
}

var dsnRE = regexp.MustCompile(`^([a-z_]+):@([a-z]+)\(([0-9a-z.]+):([0-9]+)\)$`)

func parseMariaDBDSN(dbLocation string) (dsn, error) {
	m := dsnRE.FindAllStringSubmatch(dbLocation, 1)
	if m == nil || len(m) != 1 {
		log.Printf("%#v", m[0])
		log.Printf("%#v", len(m))
		return dsn{}, fmt.Errorf("couldn't parse DSN %s", dbLocation)
	}
	user := m[0][1]
	protocol := m[0][2]
	host := m[0][3]
	port := m[0][4]
	return dsn{
		port:     port,
		host:     host,
		user:     user,
		protocol: protocol,
	}, nil
}

func printStats(stats dbapi.LexStats, validate bool) error {
	var fstr = "%-16s %6d\n"
	println("\nLEXICON STATISTICS")
	fmt.Printf(fstr, "entries", stats.Entries)
	for _, s2f := range stats.StatusFrequencies {
		//fs := strings.Split(s2f, "\t")
		//if len(fs) != 2 {
		//	return fmt.Errorf("couldn't parse status-freq from string: %s", s2f)
		//}
		//var freq, err = strconv.ParseInt(fs[1], 10, 64)
		//if err != nil {
		//	return fmt.Errorf("couldn't parse status-freq from string: %s", s2f)
		//}
		var status = "status:" + s2f.Status
		var freq = s2f.Freq
		fmt.Printf(fstr, status, freq)
	}
	if validate {
		fmt.Printf(fstr, "invalid entries", stats.ValStats.InvalidEntries)
		fmt.Printf(fstr, "validation msgs", stats.ValStats.TotalValidations)
	}
	return nil
}

func main() {
	dbapi.Sqlite3WithRegex()

	var cmdName = "importSql"

	var engineFlag = flag.String("db_engine", "sqlite", "db engine (sqlite or mariadb)")
	var dbLocation = flag.String("db_location", "", "db location (folder for sqlite; address for mariadb)")
	var dbName = flag.String("db_name", "", "db name")

	var fatalError = false
	var dieIfEmptyFlag = func(name string, val *string) {
		if *val == "" {
			fmt.Fprintln(os.Stderr, fmt.Errorf("[%s] flag %s is required", cmdName, name))
			fatalError = true
		}
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `USAGE:
      importSql [FLAGS] <SQL DUMP FILE>

      <SQL DUMP FILE> - sql dump of a lexicon database (.sql or .sql.gz)
     
     SAMPLE INVOCATIONS:
       importSql go run . -db_engine mariadb -db_location 'speechoid:@tcp(127.0.0.1:3306)' -db_name sv_db swe030224NST.pron-ws.utf8.mariadb.sql.gz
       importSql go run . -db_engine sqlite -db_location ~/wikispeech -db_name sv_db swe030224NST.pron-ws.utf8.mariadb.sql.gz

`)
		flag.PrintDefaults()

		// <NEW DB FILE>   - new (non-existing) db file to import into (<DBNAME>.db)
	}

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("?", flag.Args())
		flag.Usage()
		os.Exit(1)
	}

	dieIfEmptyFlag("db_engine", engineFlag)
	dieIfEmptyFlag("db_name", dbName)
	if fatalError {
		fmt.Fprintln(os.Stderr, fmt.Errorf("[%s] exit from unrecoverable errors", cmdName))
		os.Exit(1)
	}

	var sqlDumpFile = flag.Args()[0]

	var dbm *dbapi.DBManager
	if *engineFlag == "mariadb" {
		dbm = dbapi.NewMariaDBManager()
	} else if *engineFlag == "sqlite" {
		dbm = dbapi.NewSqliteDBManager()
	} else {
		fmt.Fprintf(os.Stderr, "invalid db engine : %s\n", *engineFlag)
		os.Exit(1)
	}

	dieIfEmptyFlag("db_engine", engineFlag)
	dieIfEmptyFlag("db_location", dbLocation)
	if fatalError {
		fmt.Fprintln(os.Stderr, fmt.Errorf("[%s] exit from unrecoverable errors", cmdName))
		os.Exit(1)
	}

	log.Printf("Input file: %s\n", sqlDumpFile)
	log.Printf("Output db: %s\n", *dbName)

	dbExists, err := dbm.DBExists(*dbLocation, lex.DBRef(*dbName))
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("[%s] %v", cmdName, err))
		os.Exit(1)
	}
	if dbExists {
		log.Fatalf("Cannot import sql dump into pre-existing database. Db already exists: %s\n", *dbName)
	}

	// _, err := os.Stat(sqlDumpFile)
	// if err != nil {
	// 	log.Fatalf("Input file does not exist : %v\n", err)
	// }

	// if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
	// 	log.Fatalf("Cannot import sql dump into pre-existing database. Db file already exists: %s\n", dbFile)
	// }

	if dbm.Engine() == dbapi.Sqlite {
		execPath := sqlitePath
		dbFile := path.Join(*dbLocation, *dbName+".db")
		/* #nosec G204 */
		cmd := exec.Command(execPath, dbFile)
		stdin := sqlDumpFile
		cmd.Stdin, err = getFileReader(stdin)
		if err != nil {
			log.Fatalf("Couldn't load sql dump %s into db %s : %v\n", sqlDumpFile, *dbName, err)
		}
		var cmdOut bytes.Buffer
		cmd.Stdout = &cmdOut
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if len(cmdOut.String()) > 0 {
			log.Println(cmdOut.String())
		}
		if err != nil {
			log.Fatalf("Couldn't load sql dump %s into db %s : %v\n", sqlDumpFile, *dbName, err)
		}

	} else if dbm.Engine() == dbapi.MariaDB {
		dsnParsed, err := parseMariaDBDSN(*dbLocation)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Parsed MariaDB DSN: %v", dsnParsed)
		// - LOAD:  gunzip -c <dumpFile> | mysql -u speechoid -h <dbHost> <dbName>
		execPath := mariaDBPath
		/* #nosec G204 */
		cmd := exec.Command(execPath, "-u", dsnParsed.user, "-h", dsnParsed.host, "--port", dsnParsed.port, "--protocol", dsnParsed.protocol, "--database", *dbName)
		stdin := sqlDumpFile
		cmd.Stdin, err = getFileReader(stdin)
		if err != nil {
			log.Fatalf("Couldn't load sql dump %s into db %s : %v\n", sqlDumpFile, *dbName, err)
		}
		var cmdOut bytes.Buffer
		cmd.Stdout = &cmdOut
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if len(cmdOut.String()) > 0 {
			log.Println(cmdOut.String())
		}
		if err != nil {
			log.Fatalf("Couldn't load sql dump %s into db %s : %v\n", sqlDumpFile, *dbName, err)
		}
	}
	log.Printf("Imported %s into db %s\n", sqlDumpFile, *dbName)

	runPostTests(dbm, *dbLocation, lex.DBRef(*dbName), sqlDumpFile)
}
