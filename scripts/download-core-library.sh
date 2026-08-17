#!/bin/sh
set -eu

version_file=${HUGINN_CORE_VERSION_FILE:-core-library.version}
version=${HUGINN_CORE_VERSION:-$(tr -d '[:space:]' < "$version_file")}
repository=${HUGINN_CORE_REPOSITORY:-killbane1232/huginn-messenger}
release_base=${HUGINN_CORE_RELEASE_BASE:-https://github.com/$repository/releases}
output_dir=${HUGINN_CORE_OUTPUT_DIR:-build}
arch=${HUGINN_CORE_ARCH:-}

if [ -z "$version" ]; then
    echo "Huginn core version is empty" >&2
    exit 1
fi

if [ -z "$arch" ]; then
    case "$(uname -m)" in
        x86_64) arch=amd64 ;;
        aarch64|arm64) arch=arm64 ;;
        *)
            echo "Unsupported Huginn core architecture: $(uname -m)" >&2
            exit 1
            ;;
    esac
fi

case "$arch" in
    amd64|arm64) ;;
    *)
        echo "Unsupported Huginn core architecture: $arch" >&2
        exit 1
        ;;
esac

asset="huginn-messenger_${version}_linux_${arch}.tar.gz"
download_url="$release_base/download/$version"
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT HUP INT TERM

curl --fail --location --silent --show-error \
    "$download_url/$asset" \
    --output "$temp_dir/$asset"
curl --fail --location --silent --show-error \
    "$download_url/SHA256SUMS" \
    --output "$temp_dir/SHA256SUMS"

grep "  $asset\$" "$temp_dir/SHA256SUMS" > "$temp_dir/SHA256SUMS.asset"
(cd "$temp_dir" && sha256sum --check SHA256SUMS.asset)

mkdir -p "$output_dir"
tar -xzf "$temp_dir/$asset" -C "$output_dir"
test -f "$output_dir/libhuginn_messenger.so"
test -f "$output_dir/libhuginn_messenger.h"
