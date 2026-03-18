#!/bin/bash 

CMD=`basename $0`

if [ $# -ne 2 ]; then
    echo "[$CMD] Use args: <SHORTNAME> <PID>"
    exit 1
fi

shortname=$1
pid=$2

pid_running=`ps -A | sed 's/^ *//' | cut -f1 -d' ' | egrep -c "^$pid$"`

if [ $pid_running -eq 0 ]; then
    echo "[$CMD] process is not running: <$shortname|PID=$pid>" >&2
    exit 1
else
    if kill -9 $pid ; then
	echo "[$CMD] killed process <$shortname|PID=$pid>" >&2
    else
	echo "[$CMD] couldn't kill process <$shortname|PID=$pid>" >&2
	exit 1	
    fi
fi
