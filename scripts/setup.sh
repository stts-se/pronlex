#!/bin/bash 

set -e

CMD=`basename $0`
SCRIPTDIR=`dirname $0`
PRONLEXPATH=`readlink -f $0 | xargs dirname | xargs dirname`
GOBINARIES=0

DBENGINE="sqlite"

print_help() {
    echo "
USAGE: bash $CMD [options]

Options:
  -h print help/options and exit
  -e db engine (optional, default: sqlite)
  -a application folder (required)
  -l db location (required for mariadb; for sqlite default is application folder)
  -b use go binaries (optional, as opposed to 'go run' with source code)

EXAMPLE INVOCATIONS:
 bash $0 -a ~/wikispeech/sqlite -e sqlite
 bash $0 -a ~/wikispeech/mariadb -e mariadb -l 'speechoid:@tcp(127.0.0.1:3306)'

Setup files will be added to the application folder.
" >&2
}


while getopts "hbe:l:a:" opt; do
    case $opt in
	h)
	    print_help
	    exit 1
	    ;;
	e)
	    DBENGINE=$OPTARG
	    if [ "$DBENGINE" != "sqlite" ] && [ "$DBENGINE" != "mariadb" ] ; then
		echo "[$CMD] Invalid db engine: $DBENGINE" >&2
		print_help
		exit 1
	    fi
	    ;;
	a)
	    APPDIR=$OPTARG
	    ;;
	l)
	    DBLOCATION=$OPTARG
	    ;;
	b)
	    GOBINARIES=1
	    ;;
	\?)
	    echo "Invalid option: -$OPTARG" >&2
	    exit 1
	    ;;
    esac
done

if [ -z "$DBENGINE" ] ; then
    echo "[$CMD] db engine must be specified using -e" >&2
    print_help
    exit 1
fi
if [ -z "$APPDIR" ] ; then
    echo "[$CMD] application folder must be specified using -a" >&2
    print_help
    exit 1
fi
if [ -z "$DBLOCATION" ] ; then
    if [ $DBENGINE == "sqlite" ]; then
	DBLOCATION=$APPDIR
    else
	echo "[$CMD] db location must be specified using -l" >&2
	print_help
	exit 1
    fi
fi

shift $(expr $OPTIND - 1 )


if [ $# -ne 0 ]; then
    echo "[$CMD] Invalid option(s): $*" >&2
    exit 1
fi


echo "[$CMD] OPTIONS:" >&2
echo "[$CMD] application folder: $APPDIR" >&2
echo "[$CMD] db engine: $DBENGINE" >&2
echo "[$CMD] db location: $DBLOCATION" >&2
echo "[$CMD] go binaries: $GOBINARIES" >&2

LEXDB=wikispeech_lexserver_demo
SS_FILES="$APPDIR/symbol_sets"
DEMOFILES=$PRONLEXPATH/lexserver/demo_files
CMDDIR="$PRONLEXPATH/cmd/lexio"

initial_setup() {
    
    if [ -f $SS_FILES ]; then
	echo "[$CMD] Application folder $APPDIR is already configured. No setup needed." >&2
	exit 0
    fi
    if [ -f $APPDIR/wikispeech_lexserver_testdb.db ] ; then
	echo "[$CMD] Application folder $APPDIR is already configured. No setup needed." >&2
	exit 0
    fi
    if [ -f $APPDIR/${LEXDB}.db ] ; then
	echo "[$CMD] Application folder $APPDIR is already configured. No setup needed." >&2
	exit 0
    fi
    # TODO: same checks for mariadb $LEXDB?


    ### LEXDATA PREPS

    echo "[$CMD] Setting up basic files in folder $APPDIR... " >&2

    mkdir -p $APPDIR || exit 1
    mkdir -p $SS_FILES || exit 1
    #mkdir -p $APPDIR/db_files || exit 1

    cp $DEMOFILES/*.sym $SS_FILES/ || exit 1
    cp $DEMOFILES/*.cnv $SS_FILES/ || exit 1
    cp $DEMOFILES/mappers.txt $SS_FILES/ || exit 1

    cp -r $PRONLEXPATH/lexserver/static $APPDIR || exit 1

    cp $PRONLEXPATH/scripts/import.sh $APPDIR || exit 1
    cp $PRONLEXPATH/scripts/start_server.sh $APPDIR || exit 1

    if [ "$DBENGINE" == "sqlite" ]; then
	DBLOCATION=`readlink -f $DBLOCATION`
    fi
    
}

function run_go_command {
    cmd=$1
    args=${@:2}
    if [ $GOBINARIES -eq 1 ]; then
	$cmd $args
    else
	go run $CMDDIR/$cmd/$cmd.go $args
    fi
}

### INITIAL SETUP
initial_setup


### LEXDATA PATHS

# scripts/.. is the root folder
cd $SCRIPTDIR/..

function parse_MariaDB_DSN() {
    ## 'USER:@PROTOCOL(HOST:PORT)
    dsn=$1;
    dsnre="^([a-z_]+):@[a-z]*\(([0-9a-z.]+):([0-9]+)\)$";
    if [[ $dsn =~ $dnsre ]]; then
	MARIADB_USER=`echo $dsn | sed 's/^\([a-z_]*\):@\([a-z]*\)(\([0-9a-z.]*\):\([0-9]*\))$/\1/'`
	MARIADB_PROTOCOL=`echo $dsn | sed 's/^\([a-z_]*\):@\([a-z]*\)(\([0-9a-z.]*\):\([0-9]*\))$/\2/'`
	MARIADB_HOST=`echo $dsn | sed 's/^\([a-z_]*\):@\([a-z]*\)(\([0-9a-z.]*\):\([0-9]*\))$/\3/'`
	MARIADB_PORT_NO=`echo $dsn | sed 's/^\([a-z_]*\):@\([a-z]*\)(\([0-9a-z.]*\):\([0-9]*\))$/\4/'`

	if [ -z $MARIADB_USER ] ||[ -z $MARIADB_PROTOCOL ] || [ -z $MARIADB_HOST ]; then # || [ -z $MARIADB_PORT ]; then
	    echo "[$CMD] Couldn't parse MariaDB DSN: $dsn" >&2
	    exit 1
	fi
	
	if [ $MARIADB_HOST == "127.0.0.1" ]; then
	    MARIADB_HOST="localhost"
	fi

	MARIADB_PORT="--port $MARIADB_PORT_NO"
	if [ -z $MARIADB_PORT_NO ]; then
	   MARIADB_PORT=""
	elif [ $MARIADB_PORT_NO == "3306" ]; then
	   MARIADB_PORT=""
	fi
	echo "MariaDB DSN $dsn parsed to user: $MARIADB_USER; host: $MARIADB_HOST" >&2
    else
	echo "[$CMD] Invalid DSN for MariaDB: $dsn" >&2
	exit 1
    fi
}

DEFAULT_MARIADB_LOCATION="speechoid:@tcp(127.0.0.1:3306)"
if [ $DBENGINE == "mariadb" ]; then
    parse_MariaDB_DSN $DBLOCATION
    # if [ $DBLOCATION != $DEFAULT_MARIADB_LOCATION ]; then
    # 	echo "[$CMD] Not not implemented for $DBENGINE location '$DBLOCATION'. Please use '$DEFAULT_MARIADB_LOCATION' or contact a developer to update this script." >&2
    # 	exit 1
    # fi
    sudo mysql -h $MARIADB_HOST $MARIADB_PORT -u root < $SCRIPTDIR/mariadb_setup.sql
    sudo mysql -h $MARIADB_HOST $MARIADB_PORT -u root -e "create database $LEXDB ; GRANT ALL PRIVILEGES ON $LEXDB.* TO '$MARIADB_USER'@'localhost' "
fi
if run_go_command createEmptyDB -db_engine $DBENGINE -db_location $DBLOCATION -db_name $LEXDB ; then
    echo "[$CMD] Created empty db $LEXDB @ $DBLOCATION for $DBENGINE" >&2
else
    echo "[$CMD] couldn't create empty db $LEXDB @ $DBLOCATION for $DBENGINE" >&2
    exit 1
fi



### LEXDATA IMPORT

echo "" >&2
echo "IMPORT: $LEXDB:sv" >&2

if run_go_command importLex -createlex -db_engine $DBENGINE -db_location $DBLOCATION -db_name $LEXDB -lex_name sv -locale sv_SE -lex_file $DEMOFILES/swe030224NST.pron-ws.utf8.for_testing -symbolset $SS_FILES/sv-se_ws-sampa.sym ; then
    echo -n ""
else
    echo "$LEXDB:sv FAILED" >&2
    exit 1
fi

echo "" >&2
if run_go_command importLex -createlex -db_engine $DBENGINE -db_location $DBLOCATION -db_name $LEXDB -lex_name nb -locale nb_NO -lex_file $DEMOFILES/nor030224NST.pron-ws.utf8.for_testing -symbolset $SS_FILES/nb-no_ws-sampa.sym ; then
    echo -n ""
else
    echo "$LEXDB:nb FAILED" >&2
    exit 1
fi

echo "" >&2
echo "IMPORT: $LEXDB:enu" >&2
if run_go_command importLex -createlex -db_engine $DBENGINE -db_location $DBLOCATION -db_name $LEXDB -lex_name enu -locale en_US -lex_file $DEMOFILES/cmudict-0.7b-ws.utf8.for_testing -symbolset $SS_FILES/en-us_ws-sampa.sym ; then
    echo -n ""
else
    echo "$LEXDB:enu FAILED" >&2
    exit 1
    fi

echo "" >&2
echo "IMPORT: $LEXDB:ar" >&2
if  run_go_command importLex -createlex -db_engine $DBENGINE -db_location $DBLOCATION -db_name $LEXDB -lex_name ar -locale ar_AR -lex_file $DEMOFILES/ar_TEST.pron-ws.utf8.for_testing -symbolset $SS_FILES/ar_ws-sampa.sym ; then
    echo -n ""
else
    echo "$LEXDB:ar FAILED" >&2
    exit 1
fi

echo "[$CMD] Clearing lexdata cache" >&2
rm -fr $APPDIR/lexdata.git


## INVOCATION TIPS
if [ $DBENGINE == "sqlite" ]; then
    invocation_args="-a $APPDIR -e $DBENGINE"
    else
    invocation_args="-a $APPDIR -e $DBENGINE -l '$DBLOCATION'"
fi

echo "
BUILD COMPLETED! YOU CAN NOW START THE LEXICON SERVER BY INVOKING:
  $ bash $SCRIPTDIR/start_server.sh $invocation_args

  USAGE INFO:
  $ bash $SCRIPTDIR/start_server.sh -h


OR IMPORT STANDARD LEXICON DATA FROM MASTER BRANCH:
  $ bash $SCRIPTDIR/import.sh $invocation_args -f <LEXDATA-GIT>

  USAGE INFO:
  $ bash $SCRIPTDIR/import.sh -h

" >&2
