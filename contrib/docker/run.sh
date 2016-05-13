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
cat > config.yaml <<EOF
addr: ${SMSHINETD_ADDR}
username: ${SMSHINETD_USERNAME}
password: ${SMSHINETD_PASSWORD}
EOF

exec ./smshinetd -logtostderr=true -stderrthreshold=INFO \
    -addr="${ADDR}" -pool="${THREADPOOL}" $*
