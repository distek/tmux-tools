#!/bin/bash

tag=$(git tag | tail -n1)

platforms=(linux darwin)
arches=(amd64 arm64)

for os in ${platforms[@]}; do
	for arch in ${arches[@]}; do
		echo "Building: platform=${os} arch=${arch}"
		GOOS="${os}" GOARCH="${arch}" go build .
		tar cvzf tmux-tools_${os}_${arch}_${tag}.tar.gz tmux-tools
		rm tmux-tools
	done
done
