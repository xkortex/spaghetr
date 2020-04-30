#!/usr/bin/env bash

# pip install . before running this script
spaghetr.server ls status_bar_dummy &
PID_SERVER=$!

sleep 0.1
spaghetr.client -health
sleep 0.1
spaghetr.client -health
sleep 0.1
spaghetr.client -health
kill $PID_SERVER

