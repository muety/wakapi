#!/bin/bash

POLL_INTERVAL=1  # sec
MEMORY_THRESHOLD=1024  # MB
TARGET_PROCESS_NAME="wakapi/wakapi"  # process under suspicion

# MySQL credentials
MYSQL_USER="wakapi"
MYSQL_PASSWORD="sshhhh"
MYSQL_HOST="127.0.01"
MYSQL_PORT="3306"

OUTPUT_DIR="/tmp"

check_free_memory() {
    FREE_MEMORY=$(free -m | awk 'NR==2{printf "%.2f\n", $7}')
    # echo "Free Memory: $FREE_MEMORY MB"
    if (( $(echo "$FREE_MEMORY < $MEMORY_THRESHOLD" | bc -l) )); then
        TARGET_PROCESS_MEM=$(ps aux | grep "$TARGET_PROCESS_NAME" | grep -v "grep" | awk '{print $6}')
        echo "[$(date)] Available memory dropped below threshold. Is now $FREE_MEMORY. Suspicious process uses $TARGET_PROCESS_MEM bytes. Dumping MySQL process list ..."
        dump_mysql_processlist
        create_heap_dump
    fi
}

dump_mysql_processlist() {
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    OUTPUT_FILE="$OUTPUT_DIR/mysql_processlist_$TIMESTAMP.txt"
    mysql -u $MYSQL_USER -p$MYSQL_PASSWORD -h $MYSQL_HOST -P $MYSQL_PORT -e "SHOW PROCESSLIST;" > $OUTPUT_FILE
    echo "[$(date)] Saved to $OUTPUT_FILE"
}

create_heap_dump() {
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    OUTPUT_FILE_1="$OUTPUT_DIR/heap_inuse_$TIMESTAMP.svg"
    OUTPUT_FILE_2="$OUTPUT_DIR/heap_alloc_$TIMESTAMP.svg"
    go tool pprof -inuse_space -svg http://localhost:6060/debug/pprof/heap > "$OUTPUT_FILE_1"
    go tool pprof -alloc_space -svg http://localhost:6060/debug/pprof/heap > "$OUTPUT_FILE_2"
    echo "[$(date)] Saved heap graph to $OUTPUT_FILE"
}

while true; do
    check_free_memory
    sleep $POLL_INTERVAL
done