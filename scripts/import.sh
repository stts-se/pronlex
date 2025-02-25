#!/bin/bash 

set -e

CMD=`basename $0`
PRONLEXPATH=`readlink -f $0 | xargs dirname | xargs dirname`
GOBINARIES=0
RELEASETAG="master"

DBENGINE="sqlite"

print_help(){
	    echo "
USAGE bash $CMD [options]

Options:
  -h print help/options and exit
  -e db engine (optional, default: sqlite)
  -a application folder (required)
  -l db location (required for mariadb; for sqlite default is application folder)
  -f lexdata folder (required)
  -r lexdata release tag (default: master)
  -b use go binaries (optional, as opposed to 'go run' with source code)

Imports lexicon data for Swedish, Norwegian, US English, and a small set of test data for Arabic from the wikispeech-lexdata repository.
Imports from sql dump files (file extension .sql.gz).
If the lexdata folder doesn't exist, it will be downloaded from github: https://github.com/stts-se/wikispeech-lexdata 

If you don't know what release tag you should use, you should probably use master.

EXAMPLE INVOCATIONS:
 bash $0 -a ~/wikispeech/sqlite -e sqlite -f ~/git_repos/wikispeech-lexdata
 bash $0 -a ~/wikispeech/mariadb -e mariadb -l 'speechoid:@tcp(127.0.0.1:3306)' -f ~/git_repos/wikispeech-lexdata
" >&2
}

while getopts "hbl:e:a:r:f:" opt; do
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
	r)
	    RELEASETAG=$OPTARG
	    ;;
	f)
	    LEXDATA=$OPTARG
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
if [ -z "$LEXDATA" ] ; then
    echo "[$CMD] lexdata folder must be specified using -f" >&2
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

if [ "$DBENGINE" == "sqlite" ]; then
    DBLOCATION=`readlink -f $DBLOCATION`
fi


echo "[$CMD] OPTIONS:" >&2
echo "[$CMD] application folder: $APPDIR" >&2
echo "[$CMD] db engine: $DBENGINE" >&2
echo "[$CMD] db location: $DBLOCATION" >&2
echo "[$CMD] lexdata folder: $LEXDATA" >&2
echo "[$CMD] lexdata release: $RELEASETAG" >&2
echo "[$CMD] go binaries: $GOBINARIES" >&2

#APPDIR=`readlink -f $2`

### LEXDATA SETUP

if [ ! -d $APPDIR/symbol_sets ] ; then
	echo "[$CMD] The application folder is not configured for the lexserver. Run setup.sh first!" >&2
    exit 1
fi

if [ -d $LEXDATA ]; then
    cd $LEXDATA && git pull && git checkout $RELEASETAG && cd - || exit 1
else
    git clone https://github.com/stts-se/wikispeech-lexdata.git $LEXDATA
    cd $LEXDATA
    git checkout $RELEASETAG
    cd -
fi

mkdir -p $APPDIR/symbol_sets || exit 1

### LEXDATA IMPORT

SVLEX=sv_se_nst_lex
NOBLEX=no_nob_nst_lex
AMELEX=en_am_cmu_lex
ARLEX=ar_ar_tst_lex

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

if [ $DBENGINE == "sqlite" ]; then
    if [ -e $APPDIR/${SVLEX}.db ]; then
	echo "[$CMD] cannot create db if it already exists: $SVLEX" >&2
	exit 1
    fi
    if [ -e $DBLOCATION/${NOBLEX}.db ]; then
	echo "[$CMD] cannot create db if it already exists: $NOBLEX" >&2
	exit 1
    fi
    if [ -e $DBLOCATION/${AMELEX}.db ]; then
	echo "[$CMD] cannot create db if it already exists: $AMELEX" >&2
	exit 1
    fi
    if [ -e $DBLOCATION/${ARLEX}.db ]; then
	echo "[$CMD] cannot create db if it already exists: $ARLEX" >&2
	exit 1
    fi
elif [ $DBENGINE == "mariadb" ]; then
    parse_MariaDB_DSN $DBLOCATION
    # DEFAULT_MARIADB_LOCATION="$MARIADB_USER:@$MARIADB_PROTOCOL($MARIADB_HOST:$MARIADB_PORT)"
    # if [ $DBLOCATION != $DEFAULT_MARIADB_LOCATION ]; then
    # 	echo "[$CMD] Not not implemented for $DBENGINE location '$DBLOCATION'. Please use '$DEFAULT_MARIADB_LOCATION' or contact a developer to update this script." >&2
    # 	exit 1
    # fi
    sudo mysql -h $MARIADB_HOST $MARIADB_PORT -u root -e "create database $SVLEX ; GRANT ALL PRIVILEGES ON $SVLEX.* TO '$MARIADB_USER'@'localhost' "
    sudo mysql -h $MARIADB_HOST $MARIADB_PORT -u root -e "create database $NOBLEX ; GRANT ALL PRIVILEGES ON $NOBLEX.* TO '$MARIADB_USER'@'localhost' "
    sudo mysql -h $MARIADB_HOST $MARIADB_PORT -u root -e "create database $AMELEX ; GRANT ALL PRIVILEGES ON $AMELEX.* TO '$MARIADB_USER'@'localhost' "
    sudo mysql -h $MARIADB_HOST $MARIADB_PORT -u root -e "create database $ARLEX ; GRANT ALL PRIVILEGES ON $ARLEX.* TO '$MARIADB_USER'@'localhost' "
fi


### COPY REQUIRED FILES
cp $LEXDATA/*/*/*.sym $APPDIR/symbol_sets/ || exit 1
echo "" >> $APPDIR/symbol_sets/mappers.txt || exit 1
cat $LEXDATA/mappers.txt >> $APPDIR/symbol_sets/mappers.txt || exit 1
cp $LEXDATA/converters/*.cnv $APPDIR/symbol_sets/ || exit 1



CMDDIR="$PRONLEXPATH/cmd/lexio"


function run_go_command {
    cmd=$1
    args=${@:2}
    if [ $GOBINARIES -eq 1 ]; then
	$cmd $args
    else
	#echo "go run $CMDDIR/$cmd/$cmd.go $args"
	go run $CMDDIR/$cmd/$cmd.go $args
    fi
}

function import_sql() {
    dbName=$1
    sqlFilePrefix=$2
    sqlFile="${sqlFilePrefix}.${DBENGINE}.sql.gz"
    # if [ $DBENGINE == "sqlite" ]; then
    # 	argz="-sqlite_folder $DBLOCATION"
    # else
    # 	argz="-mariadb_host $MARIADB_HOST -mariadb_user $MARIADB_USER -mariadb_protocol $MARIADB_PROTOCOL -mariadb_port $MARIADB_PORT"
    # fi
    if run_go_command importSql -db_engine $DBENGINE -db_location $DBLOCATION -db_name $dbName $sqlFile; then
	echo -n ""
    else
	echo "$dbName FAILED" >&2
	exit 1
    fi
}

echo "" >&2
echo "IMPORT: $SVLEX" >&2
import_sql $SVLEX $LEXDATA/sv-se/nst/swe030224NST.pron-ws.utf8

echo "" >&2
echo "IMPORT: $NOBLEX" >&2
import_sql $NOBLEX $LEXDATA/nb-no/nst/nor030224NST.pron-ws.utf8

echo "" >&2
echo "IMPORT: $AMELEX" >&2
import_sql $AMELEX $LEXDATA/en-us/cmudict/cmudict-0.7b-ws.utf8

echo "" >&2
echo "IMPORT: $ARLEX" >&2
import_sql $ARLEX $LEXDATA/ar/TEST/ar_TEST.pron-ws.utf8
