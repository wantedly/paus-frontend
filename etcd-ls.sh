#!/bin/sh

set -e

export ETCDCTL_ENDPOINT=http://`ip route show 0.0.0.0/0 | grep -Eo 'via \S+' | awk '{ print $2 }'`:2379

etcdctl ls --sort /vulcand/frontends
