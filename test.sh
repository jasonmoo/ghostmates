#!/bin/bash

set -e
set -o pipefail

COVER_FILE=/tmp/`basename $(pwd)`_`date +%s`.cov

go test -v -cover -covermode=count -coverprofile=$COVER_FILE 2>&1 | tee test.log

go tool cover -func=$COVER_FILE 2>&1 | tee -a test.log

rm -f $COVER_FILE