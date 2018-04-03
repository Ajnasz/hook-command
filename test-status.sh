#!/bin/sh

id=$(curl -s -X POST -d '{"env": {"VALAMI": "SOMETHING"}}' -H 'x-hook-token:aaa112345678' -H 'x-hook-job:example' localhost:10292);
while true; do
	echo -n "."
	test "" != "$(curl -s -X GET -H 'x-hook-token:aaa112345678' -H 'x-hook-job:example' localhost:10292/job/$id | jq -r '.info[-1]' | grep msg=EOL)" && \
		break;
	sleep 1;
done;

clear

test "" != "$(curl -s -X GET -H 'x-hook-token:aaa112345678' -H 'x-hook-job:example' localhost:10292/job/$id | jq -r '.error[-1]')" && \
	curl -s -X GET -H 'x-hook-token:aaa112345678' -H 'x-hook-job:example' localhost:10292/job/$id | jq -r '.error[]' && \
	exit 1 || \
	echo "Done $id"
