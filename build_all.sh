#!/bin/bash

# Name of the output binary
APP_NAME="txt2png"
BUILD_DIR="build"

# Define platforms and architectures
PLATFORMS=("linux" "windows" "darwin")
ARCHS=("amd64" "arm64")

# Create build directory if it doesn't exist
mkdir -p $BUILD_DIR

echo "Starting cross-platform builds..."

for os in "${PLATFORMS[@]}"; do
    for arch in "${ARCHS[@]}"; do
        output_name="${APP_NAME}-${os}-${arch}"
        if [ "$os" == "windows" ]; then
            output_name="${output_name}.exe"
        fi

        echo "Building for $os/$arch..."
        GOOS=$os GOARCH=$arch go build -o "${BUILD_DIR}/${output_name}" txt2png.go

        if [ $? -ne 0 ]; then
            echo "Error building for $os/$arch"
        fi
    done
done

echo "Builds complete! Binaries are in the '${BUILD_DIR}' directory."
ls -lh $BUILD_DIR
