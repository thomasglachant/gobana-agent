#!/bin/bash

TYPE=$1
LEVEL=$2
NUMBER=$3
FILENAME=$4

CUR_DATE=$(date +"%Y-%m-%dT%H:%M:%S.0000001+02:00")

if test -z "${TYPE}"; then
		echo "Error: invalid argument 1 : must contain log type (json|regular)"
		exit 1
fi

if test -z ${LEVEL}; then
		echo "Error: invalid argument 2 : must contain log level"
		exit 1
fi

if test -z "${NUMBER}"; then
		echo "Error: invalid argument 3 : must contain log number"
		exit 1
fi

if test -z "${FILENAME}"; then
		echo "Error: invalid argument 3 : must contain filename"
		exit 1
fi

for i in $(seq 1 "${NUMBER}"); do
	if test "${TYPE}" = "json"; then
		echo '{"message":"msg","level":400,"context":{"filename": "toto.php"}, "level_name":"'"${LEVEL}"'","channel":"request","datetime":"'${CUR_DATE}'","extra":{}}'>> "${FILENAME}"
	elif test "${TYPE}" = "symfony"; then
		echo "[${CUR_DATE}] request.${LEVEL}: My message []" >> "${FILENAME}"
	else
		echo "Error: invalid type \"${TYPE}\""
		exit 1
	fi
done

echo "OK"
