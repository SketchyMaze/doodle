#!/bin/bash

if [[ ! -d "./venv" ]]; then
	echo Creating Python virtualenv...
	python3 -m venv ./venv
	source ./venv/bin/activate
	pip install -r requirements.txt
else
	source ./venv/bin/activate
fi

python build.py

# Copy static files in.
mkdir -p compiled/pages/res
cp -r pages/res/*.* compiled/pages/res/
