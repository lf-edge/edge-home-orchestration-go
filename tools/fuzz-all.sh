#!/bin/bash

set -e

fuzzTime=${1:-10}

files=$(grep -r --include='**_test.go' --files-with-matches 'func FuzzTest' --exclude-dir=fuzzer .)

for file in ${files}
do
    funcs=$(grep -oP 'func \K(Fuzz\w*)' $file)
    for func in ${funcs}
    do
        echo "Fuzzing $func in $file"
        parentDir=$(dirname $file)
        go test $parentDir -run=$func -fuzz=$func -fuzztime=${fuzzTime}s
    done
done
