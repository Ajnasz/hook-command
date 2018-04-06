#!/bin/sh

HOOK_TOKEN="aaa112345678"
HOST=localhost:10292
id=$(curl -s -X POST -d '{"env": {"VALAMI": "SOMETHING"}}' -H "x-hook-token:${HOOK_TOKEN}" -H 'x-hook-job:example' $HOST);
while true; do
	echo -n "."
	o=$(curl -s -X GET -H "x-hook-token:${HOOK_TOKEN}" $HOST/job/$id)
	echo $o
	test "" != "$(echo $o| grep msg=EOL)" && \
		break;
	sleep 1;
done;

# clear

test "" != "$(curl -s -X GET -H "x-hook-token:${HOOK_TOKEN}" $HOST/job/$id | grep "level=Error")" && {
		echo "Error detected $id"
		curl -s -X GET -H "x-hook-token:${HOOK_TOKEN}" $HOST/job/$id | grep "level=Error"
		exit 1
	}

echo "Done $id"
exit 0
