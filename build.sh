#!/bin/sh
rm -rf out && mkdir -p out

prog="nekoget"

[ -z "${GOPARM}" ] && export GOARM=5
[ -z "${GOAMD64}" ] && export GOAMD64=v1
[ -z "${CGO_ENABLED}" ] && export CGO_ENABLED=0
[ -z "${ESOTERIC}" ] && export ESOTERIC=0

for x in $(go tool dist list); do
	export GOOS=$(echo $x | cut -d'/' -f1)
	export GOARCH=$(echo $x | cut -d'/' -f2)

	# fuck that shit
	if [ "${ESOTERIC}" = 0 ]; then
		case "$GOOS" in
		freebsd)
			;;
		windows)
			;;
		linux)
			;;
		illumos)
			;;
		*)
			continue
			;;
		esac

		case "$GOARCH" in
		amd64)
			;;
		386)
			;;
		arm64)
			;;
		arm)
			;;
		*)
			continue
			;;
		esac
	fi

	[ "${GOARCH}" = 'wasm' ] && continue

	OUT="./out/${prog}-${GOOS}-${GOARCH}"

	if [ "${GOOS}" = 'windows' ]; then
		OUT="${OUT}.exe"
	fi

	go build -ldflags='-s -w' -trimpath -o "$OUT"
done