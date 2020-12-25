#!/bin/sh

LOG_FILES=$(ls /var/log/containers/)
ii=0

echo "["
for LOG_FILE in $LOG_FILES
do
	if [ $ii -eq 0 ]; then
		let "ii++"
	else
		echo ","
	fi
	LOG_FILE_RAW_COUNT=$(wc -l alertmanager-main-0_openshift-monitoring_alertmanager-123...789.log"$LOG_FILE" | awk 'END {print $1}')
	STRING=${LOG_FILE%-*}
	POD_NAME=$(echo $STRING | awk '{split($0,a,"_"); print a[1]}')
	NAMESPACE=$(echo $STRING | awk '{split($0,a,"_"); print a[2]}')
	CONTAINER_NAME=$(echo $STRING | awk '{split($0,a,"_"); print a[3]}')

	echo "{\"namespace\":\"$NAMESPACE\",\"pod_name\":\"$POD_NAME\",\"container_name\":\"$CONTAINER_NAME\",\"row_count\":$LOG_FILE_RAW_COUNT}"
done
echo "]"

