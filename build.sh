#!/bin/bash -e
NOW=$(date -u +'%Y-%m-%d_%TZ')
HEAD_SHA1=$(git rev-parse HEAD)
HEAD_TAG=$(git tag --points-at HEAD | grep -e "^v" | sort | tail -1 | cut -b2-)
TOOL_NAME="docker-log-watch"

echo "Building ${TOOL_NAME} version ${HEAD_TAG}..."
for platform in windows/amd64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 ; do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	BINARY_NAME="${TOOL_NAME}.${GOOS}-${GOARCH}"
	if [ "${GOOS}" = "windows" ]; then
		BINARY_NAME+='.exe'
	fi
	echo "Building for platform ${platform}: ${BINARY_NAME}"
	env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X main.versionTag=${HEAD_TAG} -X main.versionSha1=${HEAD_SHA1} -X main.buildDate=${NOW}" -o build/${BINARY_NAME}
done
