#!/bin/bash
trap "rm server; kill 0" EXIT

go build -o server
./server -port=8001 &
sleep 1
./server -port=8002 &
sleep 1
./server -port=8003 -api=1 &

sleep 2
echo ">>> star test\n"
echo ">>> test 1\n"
curl "http://localhost:9999/api?key=Lisi" &
sleep 1
echo "\n>>> test 2"
curl "http://localhost:9999/api?key=Lisi" &
sleep 1
echo "\n>>> test 3"
curl "http://localhost:9999/api?key=Lisi" &

wait
