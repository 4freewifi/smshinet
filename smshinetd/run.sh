#!/bin/sh
exec ./smshinetd -logtostderr=true -stderrthreshold=INFO $*
