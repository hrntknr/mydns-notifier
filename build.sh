#!/bin/bash

targets=(
	"darwin amd64"
	"darwin 386"
	"freebsd amd64"
	"freebsd 386"
	"linux amd64"
	"linux 386"
	"windows amd64"
	"windows 386"
)

if [ -e release ];then
	rm -r release
fi

for target in "${targets[@]}"
do
	IFS=" "
	set -- $target
	export GOOS=$1
	export GOARCH=$2
	echo "building ${GOOS}/${GOARCH}"
	if [ $GOOS = "windows" ];then
		go build -o "release/${GOOS}_${GOARCH}_mydnsnotifier.exe"
	else
		go build -o "release/${GOOS}_${GOARCH}_mydnsnotifier"
	fi
done
