#!/bin/bash

commit=`git rev-parse --short HEAD`
built_at=`date +%FT%T%z`
built_by=${USER}
built_on=`hostname`

go build -ldflags "-X main.commit=${commit} -X main.builtAt='${built_at}' -X main.builtBy=${built_by} -X main.builtOn=${built_on}"

