# pronlex
pronlex is a pronunciation lexicon database with a server behind a simple HTTP API.

[![GoDoc](https://godoc.org/github.com/stts-se/pronlex?status.svg)](https://godoc.org/github.com/stts-se/pronlex)
[![Go Report Card](https://goreportcard.com/badge/github.com/stts-se/pronlex)](https://goreportcard.com/report/github.com/stts-se/pronlex) [![Build Status](https://travis-ci.com/stts-se/pronlex.svg?branch=master)](https://app.travis-ci.com/stts-se/pronlex)



## Lexicon server / Installation instructions

Utility scripts below (setup, import, start_server) require a working `bash` installation (preferably on a Linux system).

### I. Installation

1. Prerequisites

     If you're on Linux, you may need to install `gcc` and `build-essential` for the `sqlite3` go adapter to work properly:   
     `$ sudo apt-get install gcc build-essential`

2. Set up `go`

     Download: https://golang.org/dl/ (1.24 or higher)   
     Installation instructions: https://golang.org/doc/install             


3. Install database support

   [Sqlite3](https://www.sqlite.org/): On Linux systems with `apt`, run `sudo apt install sqlite3`

   [MariaDB](https://mariadb.org/): On Linux systems with `apt`, run `sudo apt install mariadb-server` or similar (it should be version 10.1.3 or higher)

   Please note that you need to install both databases if you intend to run unit tests or other automated tests

4. Clone the source code

   `$ git clone https://github.com/stts-se/pronlex.git`  
   `$ cd pronlex`   
   
5. Test (optional)

   `pronlex$ go test ./...`

6. Set up MariaDB (optional)

   ``` sh
   sudo mysql -u root < scripts/mariadb_setup.sql`
   cd dbapi
   go test . -mariadb # run unit tests (optional)
   ```



### II. Quick start: Create a lexicon database file and look up a word (for Sqlite configuration)

1) Download an SQL lexicon dump file. In the following example, we use a Swedish lexicon: `https://github.com/stts-se/wikispeech-lexdata/blob/master/sv-se/nst/swe030224NST.pron-ws.utf8.sqlite.sql.gz`

2) Pre-compile binaries (for faster execution times)

    `pronlex$ go build ./...`

2) Create a database file (this takes a while):

    `pronlex$ importSql -db_engine sqlite -db_location ~/wikispeech/sqlite/ -db_name sv_db swe030224NST.pron-ws.utf8.sqlite.sql.gz`
       
3) Test looking up a word:
       
   `pronlex$ lexlookup -db_engine sqlite -db_location ~/wikispeech/sqlite/ -db_name sv_db -lexicon swe_lex åsna`


### III. Server setup

1. Setup the pronlex server:

   `pronlex$ bash scripts/setup.sh -a <application folder> -e <db engine> -l <db location>*`   
   Example:     
   `pronlex$ bash scripts/setup.sh -a ~/wikispeech/sqlite -e sqlite`    
   Usage info:      
   `pronlex$ bash scripts/setup.sh -h`

   Sets up the pronlex server using the specified database engine and specified location, and a set of test data. The db location folder is not required for sqlite (if it's not specified, the application folder will be used for db location).

   The application folder is where databases and other resources will be installed. It can be any folder of your choice.

   If, for some reason, you are not using the above setup script to configure your pronlex installation, you need to configure mariadb using the mariadb setup script (as root):

   `sudo mysql -u root < scripts/mariadb_setup.sql`


2. Import lexicon data (optional)

   `pronlex$ bash scripts/import.sh -a <application folder> -e <db engine> -l <db location>* -f <lexdata git> `    
   Example:      
   `pronlex$ bash scripts/import.sh -a ~/wikispeech/sqlite -e sqlite -f ~/git_repos/wikispeech-lexdata`   

   Imports lexicon databases (sql dumps) for Swedish, Norwegian, US English, and a small set of test data for Arabic from the [wikispeech-lexdata](https://github.com/stts-se/wikispeech-lexdata) repository.
If the `<lexdata git>` folder exists on disk, lexicon resources will be read from this folder. If it doesn't exist, the lexicon data will be downloaded from github.
The db location folder is not required for sqlite (if it's not specified, the application folder will be used for db location).

   If you want to import other lexicon data, or just a subset of the data above, you can use one of the following methods:
   
   * Import lexicon files from the command line: https://github.com/stts-se/pronlex/tree/master/cmd/lexio/importLex.
   * Import database sql dumps files from the command line: https://github.com/stts-se/pronlex/tree/master/cmd/lexio/importSql.


You can create your own lexicon files, or you can use existing data in the [wikispeech-lexdata](https://github.com/stts-se/wikispeech-lexdata) repository. The lexicon file format is described here: https://godoc.org/github.com/stts-se/pronlex/line.


### IV. Start the lexicon server

The server is started using this script:

`pronlex$ bash scripts/start_server.sh -e <db engine> -l <db location>* -a <application folder>`

Example: 

`pronlex$ bash scripts/start_server.sh -e sqlite -a ~/wikispeech/sqlite/`

The startup script will run some init tests in a separate test server, before starting the standard server.

When the standard (non-testing) server is started, it always creates a demo database and lexicon, containing a few simple entries for demo and testing purposes. The server can thus be started and tested even if you haven't imported the lexicon data above.

For a complete set of options, run:  
`pronlex$ bash scripts/start_server.sh -h`


<!--

## For developers

If you are developing for Wikispeech, and need to make changes to this repository, make sure you run a test build using `build_and_test.sh` before you make a pull request. Don't run more than one instance of this script at once, and make sure no pronlex server is already running on the default port.

Wikimedia's installation instructions for Wikispeech: https://www.mediawiki.org/wiki/Extension:Wikispeech

-->


---

_This work was supported by the Swedish Post and Telecom Authority (PTS) through the grant "Wikispeech – en användargenererad talsyntes på Wikipedia" (2016–2017), and the Swedish Inheritance Fund (Allmänna arvsfonden) through the grant "Wikispeech – Talsyntes och taldatainsamlare." (2024–2027)._
