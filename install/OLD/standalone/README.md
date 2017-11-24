# Standalone setup

Below are instructions on how to set up the lexicon server for standalone use. For developer setup, see the `developer` folder.

## I. Preparation steps

1. Prerequisites

     If you're on Linux, you may need to install `gcc` and `build-essential` for the `sqlite3` go adapter to work properly:   
     `$ sudo apt-get install gcc build-essential`

2. Set up `go`

     Download: https://golang.org/dl/ (1.8 or higher)  
     Installation instructions: https://golang.org/doc/install
 
        
3. Install [Sqlite3](https://www.sqlite.org/)

   On Linux systems with `apt`, run `sudo apt install sqlite3`

## II. Installation

1. Download the install script for the release you want, or get it from this README's git folder.    
   Download for the master branch: [install.sh](https://raw.githubusercontent.com/stts-se/pronlex/master/install/standalone/install.sh)

2. Install the pronlex server

     `$ bash install.sh <APPDIR>`

   Installs the pronlex server and a set of test data into the folder specified by `<APPDIR>`.


3. Import lexicon data (optional)

    `$ bash <APPDIR>/import.sh`

   Imports lexicon data for Swedish, Norwegian, US English and a small test file for Arabic.


## III. Start the lexicon server

The server is started using this script

`$ bash <APPDIR>/start_server.sh`

The startup script will run some init tests in a separate test server, before starting the standard server.

When the standard (non-testing) server is started, it always creates a demo database and lexicon, containing a few simple entries for demo and testing purposes. The server can thus be started and tested even if you haven't imported the lexicon data above.

To specify port, run:   
`$ bash <APPDIR>/start_server.sh -p <PORT>`


For a complete set of options, run:  
`$ bash <APPDIR>/start_server.sh -h`
