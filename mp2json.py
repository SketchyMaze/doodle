#!/usr/bin/env python3

"""mp2json: convert a msgpack binary file into JSON for debugging."""

import msgpack
import json
import sys

if len(sys.argv) < 2:
    print("Usage: mp2json <filename.level>")

with open(sys.argv[1], 'rb') as fh:
    header = fh.read(8)
    magic = header[:6].decode("utf-8")
    if magic != "DOODLE":
        print("input file doesn't appear to be a doodle drawing binary")
        sys.exit(1)

    reader = msgpack.Unpacker(fh, raw=False, max_buffer_size=10*1024*1024)
    for o in reader:
        print(o)
        print(json.dumps(o, indent=2))
