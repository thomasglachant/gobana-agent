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
		echo '{"message":"msg","level":400,"context":{"filename": "toto.php"}, "level_name":"'"${LEVEL}"'","channel":"request","datetime":"'${CUR_DATE}'","extra":{"app_user": "be78bf26-3714-43da-aa92-bd4a7be29d22"}}'>> "${FILENAME}"
	elif test "${TYPE}" = "json_nginx"; then
		echo '{"http.url":"/","http.version":"HTTP/1.1","http.status_code":401,"http.method":"HEAD","http.referer":"","http.useragent":"Zabbix 6.2.1","time_local":"19/Nov/2022:09:45:46 +0100","remote_addr":"51.91.147.220","remote_user":"","body_bytes_sent":"0","request_time":0.010,"response_content_type":"application/json","X-Forwarded-For":"51.91.147.220", "extra": {"username": "michel"} }'  >> "${FILENAME}"
	elif test "${TYPE}" = "symfony"; then
		echo "[${CUR_DATE}] request.${LEVEL}: My message []" >> "${FILENAME}"
	elif test "${TYPE}" = "nginx"; then
		if test "${LEVEL}" = "CRITICAL"; then
		echo '126.76.96.124 - XXXX@localhost.local [25/Jul/2022:10:49:48 +0200] "GET /status HTTP/1.1" 500 44 "https://api.localhost.local/status" "Mozilla/5.0+(compatible; UptimeRobot/2.0; http://www.uptimerobot.com/)" "126.76.96.124"' >> "${FILENAME}"
		else
		echo '126.76.96.124 - XXXX@localhost.local [25/Jul/2022:10:49:48 +0200] "GET /status HTTP/1.1" 200 44 "https://api.localhost.local/status" "Mozilla/5.0+(compatible; UptimeRobot/2.0; http://www.uptimerobot.com/)" "126.76.96.124"' >> "${FILENAME}"
		fi
	else
		echo "Error: invalid type \"${TYPE}\""
		exit 1
	fi
done

echo "OK"
