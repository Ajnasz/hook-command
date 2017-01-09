#!/bin/sh

echo "Exmaple run"
sleep 1
echo "Exmaple run 2" >&2
exit 1
sleep 1
echo "EXAMPLE ENV value is: $EXAMPLE"
