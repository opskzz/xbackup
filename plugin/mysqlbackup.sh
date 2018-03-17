#!/bin/bash
#Mysql Complete Hot Backup
#Version V2.1
##@date 2016-6-23 修改增量备份文件生成模式
##@date 2016-9-13 添加centos7兼容
##@2016-09-27 修改为xbstream的压缩备份模式
##@date 2017-04-25 添加多实例兼容参数 $2，添加iops动态判断，修改为yum源自动安装xtrabackup最新版本以兼容mysql5.7

# Source function library.
mysql_config_file=$1
. /etc/init.d/functions

PATH=$PATH:/data/mysql/bin
export PATH

# 保留备份的分钟数
MIN=1410

PORT=$(sed -n '/\[mysqld\]/,//p' ${mysql_config_file}|grep -E '^port\s*='|awk -F '=' '{print $2}'| head -1|xargs echo)
if [ -n "$2" ] ;then
    PORT=$2
fi

USER="xxx"
MYSQLHOST="127.0.0.1"
PASSWORD="xxx"
XTRABACKUP="xtrabackup"
IOPS=50
n=`cat ${mysql_config_file}|grep -E '^port\s*=\s*33\s*'|sed s/[[:space:]]//g|sort|uniq|wc -l`
if [ ${n} -gt 1 ]; then 
    IOPS=`expr 60 / ${n}`
fi

COMP_THR=4
USE_MEM="10M"
DATE=$(date '+%F_%H-%M')
HOUR=$(date +%H)
BACKUP_PATH="/data/backup/database_${PORT}"
LOG_FILE="/var/log/xtrabackup_${PORT}.log"
MYSQL_LOG="/var/log/mysql_${PORT}.log"
HOST=`ip a|grep -E "([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})" | awk '{print $2}' | cut -d "/" -f 1 | grep -E "^192\.|^10\." | head -1`
LAST_ALL_LOG="/var/log/last_${PORT}.log"

PID_FILE="/tmp/hotbak_inc_mysql_${PORT}"
LOCK_FILE="/var/lock/subsys/xtrabackup_${PORT}"

function out_log () {
    if [ -n "$1" ]; then
        _PATH="$1"
    else
        echo "unknown error"
        echo -e "wrong\nunknown error" > ${MYSQL_LOG}
        exit
    fi

    [ ! -f ${LOG_FILE} ] && touch "${LOG_FILE}"
    echo -e "[$(date +%Y-%m-%d' '%H:%M:%S)] ${_PATH}" >> "${LOG_FILE}"
}

function del_bak () {
    for dbfile in `find "${1}/" -name "*.compress.xbstream" -type f -mmin +${MIN}`; do
        out_log "delete from ${dbfile}"
        rm -f ${dbfile}
    done
}

function all_back() {
    DB_NAME=("$HOST"_"$DATE"_"$PORT")
    echo 'all'
    out_log "cd ${BACKUP_PATH}"
    cd ${BACKUP_PATH}
    
    out_log "innobackupex --defaults-file=${mysql_config_file} --throttle=${IOPS} --host=${MYSQLHOST} --port=${PORT} --user=${USER} --password=${PASSWORD} --extra-lsndir=${BACKUP_PATH} --no-lock --stream=xbstream --compress --compress-threads=${COMP_THR} ${BACKUP_PATH} > ${BACKUP_PATH}/${DB_NAME}.compress.xbstream 2>>${LOG_FILE}"
    innobackupex --defaults-file=${mysql_config_file} --throttle=${IOPS} --host=${MYSQLHOST} --port=${PORT} --user=${USER} --password=${PASSWORD} --extra-lsndir=${BACKUP_PATH} --no-lock --stream=xbstream --compress --compress-threads=${COMP_THR} ${BACKUP_PATH} > ${BACKUP_PATH}/${DB_NAME}.compress.xbstream 2>>${LOG_FILE}
    RETVAL=$?
    
    if  [ ${RETVAL} -eq 0 -a `tail -50 "${LOG_FILE}" | grep -ic "\berror\b"` -eq 0 ]; then
        del_bak "${BACKUP_PATH}"
        echo -n $"Complete Hot Backup"
        success
        echo
        echo -e "ok\n${DB_NAME}.compress.xbstream" > ${MYSQL_LOG}
        echo -e "ok\n${DB_NAME}.compress.xbstream" > ${LAST_ALL_LOG}
        rm -f ${LOCK_FILE}
    else
        out_log "[ERROR] error: xtrabackup failure"
        echo -n $"Complete Hot Backup"
        failure
        echo
        echo -e "wrong\n$(tail -50 "${LOG_FILE}" | grep -i "\berror\b" | sed -n '1p')" > ${MYSQL_LOG}
        echo -e "wrong\n${DB_NAME}.compress.xbstream" > ${LAST_ALL_LOG}
        rm -f ${LOCK_FILE}
    fi
}

function inc_back() {
    CHECKPOINT=$(awk '/to_lsn/ {print $3}' ${BACKUP_PATH}/xtrabackup_checkpoints 2>/dev/null)
    DB_NAME=("$HOST"_"$DATE"_"$PORT""-increase")
    echo "inc"
    out_log "cd ${BACKUP_PATH}"
    cd ${BACKUP_PATH}
    
    if [ ! -e "${BACKUP_PATH}/xtrabackup_checkpoints" ];then
        out_log "xtrabackup_checkpoints does not exist"
        all_back
        exit 0
    fi
    
    out_log "innobackupex --defaults-file=${mysql_config_file} --throttle=${IOPS} --host=${MYSQLHOST} --port=${PORT} --user=${USER} --password=${PASSWORD} --no-lock --incremental --incremental-lsn=${CHECKPOINT} --stream=xbstream --compress --compress-threads=${COMP_THR} ${BACKUP_PATH} > ${BACKUP_PATH}/${DB_NAME}.compress.xbstream 2>>${LOG_FILE}"
    innobackupex --defaults-file=${mysql_config_file} --throttle=${IOPS} --host=${MYSQLHOST} --port=${PORT} --user=${USER} --password=${PASSWORD} --no-lock --incremental --incremental-lsn=${CHECKPOINT} --stream=xbstream --compress --compress-threads=${COMP_THR} ${BACKUP_PATH} > ${BACKUP_PATH}/${DB_NAME}.compress.xbstream 2>>${LOG_FILE}
    RETVAL=$?
    
    if  [ ${RETVAL} -eq 0 -a `tail -50 "${LOG_FILE}" | grep -ic "\berror\b"` -eq 0 ]; then        
        del_bak "${BACKUP_PATH}"
        echo -n $"Incremental Hot Backup"
        success
        echo
        echo -e "ok\n${DB_NAME}.compress.xbstream" >${MYSQL_LOG}
        rm -f ${LOCK_FILE}
    else
        out_log "[ERROR] error: xtrabackup failure"
        echo -n $"Incremental Hot Backup"
        failure
        echo
        echo -e "wrong\n$(tail -50 "${LOG_FILE}" | grep -i "\berror\b" | sed -n '1p')" >${MYSQL_LOG}
        rm -f ${LOCK_FILE}
    fi
}

if ! rpm -qa | grep percona-xtrabackup > /dev/null 2>&1;then
    out_log "start install percona-xtrabackup"
    yum install -y percona-xtrabackup-24
fi

if ! rpm -qa | grep qpress > /dev/null 2>&1;then
    out_log "start install qpress"
    yum install -y qpress
fi

if [ ! -d "$BACKUP_PATH" ]; then
    out_log "mkdir -p ${BACKUP_PATH}"
    mkdir -p ${BACKUP_PATH}
    out_log "chown -R nobody.nobody ${BACKUP_PATH}"
    chown -R nobody.nobody ${BACKUP_PATH}
fi

[ ! -f $PID_FILE ] && touch ${PID_FILE}
_PID=`cat ${PID_FILE}`
if [ `ps ax|awk '{print $1}'|grep -v grep|grep -c "\b${_PID}\b"` -eq 1 ] && [ -f ${LOCK_FILE} ]; then
    echo -n $"xtrabackup process already exist."
    echo -e 'wrong\nxtrabackup process already exist.' >${MYSQL_LOG}
    exit
else
    echo $$ >${PID_FILE}
    touch ${LOCK_FILE}
fi

if [ "${HOUR}" -eq "4" ];then
    all_back
else
    inc_back
fi
