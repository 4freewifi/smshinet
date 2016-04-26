#!/bin/sh

set -e

for E in SMSHINETD_ADDR SMSHINETD_USERNAME SMSHINETD_PASSWORD; do
    val=$(eval echo -n x"\${$E}")
    test "${val}" = "x" && (
        echo "Please set environment variable: $E"
        exit 1
    )
done

# generate config file `config.yaml'
confd -onetime -backend=env -confdir="/go/src/github.com/4freewifi/smshinet/contrib/confd"

exec ./smshinetd -logtostderr=true -stderrthreshold=INFO \
    -addr="${ADDR}" -pool="${THREADPOOL}" -conf="config.yaml"
