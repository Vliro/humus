#!/bin/sh
cleanup() {
    killall background
    pkill dgraph
    sleep 1
    rm -rf p w zw
}

testcmd () {
    command -v "$1" >/dev/null
}
trap cleanup EXIT
trap cleanup INT

SCHEMA='type InnerValue {\nvalue7: int\n}\n type Value {\n value: string  \n value2: int \n value3: uid \n value4: [uid] \n value5: [int] \n value6: [string] \n facet: uid  \n}\n value: string @index(hash) . \n value2: int .\n value3: uid .\n value4: [uid] .\n value5: [int] .\n value6: [string] .\n facet: uid @reverse . \n value7: int .'

#Install dgraph if it doesnt exist.
if testcmd dgraph; then
    echo
else
    go get -u github.com/dgraph-io/dgraph/dgraph
fi
#init default dgraph zero
dgraph zero >/dev/null 2>&1 &
sleep 1
#start alpha at default port, GRPC 9080 & http 8080
dgraph alpha --lru_mb 2048 > /dev/null 2>&1 &
sleep 5
echo "$SCHEMA"
STATUS=$(curl -X POST localhost:8080/alter -d "$SCHEMA")
echo "$STATUS"
sleep infinity