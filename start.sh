#!/bin/sh
export PATH=/usr/local/bin/:/usr/bin:/usr/local/bin
export HCMD_CONFIGFILE=./configuration.example.json
export HCMD_CONFIGDIR=./configuration.example.d/
export HCMD_SCRIPTSDIR=./scripts
export HCMD_TOKEN=aaa112345678
./hook-command
