#!/bin/sh

set -e

go test -c
exec ./smshinet.test -test.v -logtostderr=true -stderrthreshold=INFO -v=1
