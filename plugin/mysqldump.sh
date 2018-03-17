#!/bin/bash
DAY="7"
PORT="3306"
USER="xxx"
HOST="127.0.0.1"
PASSWORD="xxx"
THREADS="12"
BACKUP_PATH="/data/backup/mydumper_${PORT}"
LOG_FILE="/var/log/mydumper_${PORT}.log"
PID_FILE="/var/lock/subsys/mydumper_${PORT}"
LOCK_FILE="/var/lock/subsys/mydumper_${PORT}"
IPADDR=$(ifconfig|awk -F':' '/inet addr:/{gsub(/[^0-9.]/,"",$2);print $2}'|egrep -v "^127"|xargs)

out_log() {
	if [ -n "$1" ]; then
		_PATH="$1"
	else
		echo "unknown error"
		echo -e "wrong\nunknown error" > ${LOG_FILE}
		exit 1
	fi
	[ ! -f ${LOG_FILE} ] && touch "${LOG_FILE}"
	echo -e "[$(date +%F' '%T)] ${_PATH}" >> "${LOG_FILE}"
}

del_backup() {
	for dbfile in `find "${1}/" -type f -mtime +${DAY}`; do
		out_log "delete from ${dbfile}"
		rm -f ${dbfile}
	done
}

if [ `which mydumper` -ne 1 ] > /dev/null 2>&1;then
    echo -e "[$(date +%F' '%T)] mydumper does't exist"
    exit 1
fi
if [ `which myloader` -ne 1 ] > /dev/null 2>&1;then
    echo -e "[$(date +%F' '%T)] myloader does't exist"
    exit 1
fi

if [ ! -d "$BACKUP_PATH" ]; then
	out_log "mkdir -p ${BACKUP_PATH}" && mkdir -p ${BACKUP_PATH}
	out_log "chown -R nobody.nobody ${BACKUP_PATH}"
	chown -R nobody.nobody ${BACKUP_PATH}
fi

[ ! -f $PID_FILE ] && touch ${PID_FILE}
_PID=`cat ${PID_FILE}`
if [ `ps aux|awk '{print $1}'|grep -v grep|grep -c "\b${_PID}\b"` -eq 1 ] && [ -f ${LOCK_FILE} ]; then
    out_log "mydumper process already exist"
    echo -n "mydumper process already exist"
    exit 1
else
    echo $$ >${PID_FILE}
    touch ${LOCK_FILE}
fi

dump_backup() {
    local NOW=$(date '+%F_%H-%M')
    local DUMPDIR=$BACKUP_PATH/$NOW
    mkdir -p $DUMPDIR
    out_log "mkdir -p $DUMPDIR"
    out_log "mydumper -u $USER -p $PASSWORD -P $PORT --threads=$THREADS -e --regex '^(?!(mysql|test|performance_schema|information_schema))' --compress-protocol -o $DUMPDIR"
    mydumper -u $USER -p $PASSWORD -P $PORT --threads=$THREADS -e --regex '^(?!(mysql|test|performance_schema|information_schema))' --compress-protocol -o $DUMPDIR
    RETVAL=$?
    if [ $RETVAL -eq 0 ];then
        del_backup "$BACKUP_PATH"
        echo -n "Complete Dump Backup Success"
        rm -f $LOCK_FILE
    else
        echo -n "Complete Hot Backup Faild"
        rm -f ${LOCK_FILE}
        exit 1
    fi
    cd $BACKUP_PATH
    tar zcf $IPADDR-$NOW.tar.gz $NOW
    if [ $? -eq 0 ];then
        out_log "tar zcf $IPADDR-$NOW.tar.gz $NOW Success"
        rm -rf $NOW && exit 0
    else
        out_log "tar zcf $IPADDR-$NOW.tar.gz $NOW Faild"
        exit 1
    fi
}

dump_backup
