#!/usr/bin/env bash

## Remove all generated files
rm -r build
rm -r dist
rm -r wheels

rm -r spaghetr/protos/*_pb2.py
rm -r spaghetr/protos/*_pb2_grpc.py