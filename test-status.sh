#!/bin/sh

HOOK_TOKEN="aaa112345678"
HOST=localhost:10292
id=$(curl -s -X POST -d '{"env": {"VALAMI": "SOMETHING"}}' -H "x-hook-token:${HOOK_TOKEN}" -H 'x-hook-job:example' $HOST)
echo "Job: $id"
while true; do
	echo -n "." >&2
	o=$(curl -s -X GET -H "x-hook-token:${HOOK_TOKEN}" $HOST/job/$id)
	test "" != "$(echo $o| grep 'msg=EOL')" && \
		break;
	sleep 1;
done;

# clear

test "" != "$(curl -s -X GET -H "x-hook-token:${HOOK_TOKEN}" $HOST/job/$id | grep "level=error")" && {
		echo >&2
		echo "Error detected" >&2
		curl -s -X GET -H "x-hook-token:${HOOK_TOKEN}" $HOST/job/$id | grep "level=error" >&2
		exit 1
	}

echo "Done"
exit 0
