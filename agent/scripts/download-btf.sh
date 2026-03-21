#!/bin/bash
# download-btf.sh - Download BTF files from btfhub-archive for Amazon Linux 2/2023
#
# Downloads pre-built BTF files for kernels that don't ship with
# /sys/kernel/btf/vmlinux, enabling CO-RE eBPF on those systems.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BTF_DST_DIR="$PROJECT_ROOT/business_plugins/ebpf_base_detector/btf"

BTFHUB_REPO="https://github.com/aquasecurity/btfhub-archive.git"
CLONE_DIR=$(mktemp -d)

# Minimum kernel version to include
MIN_MAJOR=5
MIN_MINOR=10

cleanup() {
    rm -rf "$CLONE_DIR"
}
trap cleanup EXIT

# Compare kernel version: returns 0 if $1 >= 5.10
kernel_ge_min() {
    local ver="$1"
    local major minor
    major=$(echo "$ver" | cut -d. -f1)
    minor=$(echo "$ver" | cut -d. -f2)

    if [ "$major" -gt "$MIN_MAJOR" ]; then
        return 0
    elif [ "$major" -eq "$MIN_MAJOR" ] && [ "$minor" -ge "$MIN_MINOR" ]; then
        return 0
    fi
    return 1
}

echo "=== BTF File Downloader ==="
echo "Destination: $BTF_DST_DIR"
echo ""

# Clone btfhub-archive (shallow)
echo "Cloning btfhub-archive (shallow)..."
git clone --depth=1 "$BTFHUB_REPO" "$CLONE_DIR" 2>&1 | tail -1

mkdir -p "$BTF_DST_DIR"

count=0
skipped=0
already=0

# Process Amazon Linux 2 and 2023 x86_64 BTF files
for amzn_ver in 2 2023; do
    src_dir="$CLONE_DIR/amazonlinux/$amzn_ver/x86_64"
    if [ ! -d "$src_dir" ]; then
        echo "Warning: $src_dir not found, skipping Amazon Linux $amzn_ver"
        continue
    fi

    echo ""
    echo "Processing Amazon Linux $amzn_ver (x86_64)..."

    for btf_file in "$src_dir"/*.btf; do
        [ -f "$btf_file" ] || continue

        filename=$(basename "$btf_file")
        # Extract kernel version from filename (e.g., 5.10.184-175.731.amzn2.x86_64.btf)
        kernel_ver="$filename"

        # Get major.minor from filename
        ver_prefix=$(echo "$filename" | grep -oE '^[0-9]+\.[0-9]+' || true)
        if [ -z "$ver_prefix" ]; then
            skipped=$((skipped + 1))
            continue
        fi

        if ! kernel_ge_min "$ver_prefix"; then
            skipped=$((skipped + 1))
            continue
        fi

        # Target filename: strip .btf suffix, use kernel release as name
        # btfhub filenames are already <uname-r>.btf
        dst_file="$BTF_DST_DIR/$filename"

        if [ -f "$dst_file" ]; then
            already=$((already + 1))
            continue
        fi

        cp "$btf_file" "$dst_file"
        count=$((count + 1))
    done
done

# Also handle .btf.tar.xz compressed files if present
for amzn_ver in 2 2023; do
    src_dir="$CLONE_DIR/amazonlinux/$amzn_ver/x86_64"
    [ -d "$src_dir" ] || continue

    for btf_archive in "$src_dir"/*.btf.tar.xz; do
        [ -f "$btf_archive" ] || continue

        archive_name=$(basename "$btf_archive")
        # Extract the .btf filename (remove .tar.xz)
        btf_name="${archive_name%.tar.xz}"

        ver_prefix=$(echo "$btf_name" | grep -oE '^[0-9]+\.[0-9]+' || true)
        if [ -z "$ver_prefix" ]; then
            skipped=$((skipped + 1))
            continue
        fi

        if ! kernel_ge_min "$ver_prefix"; then
            skipped=$((skipped + 1))
            continue
        fi

        dst_file="$BTF_DST_DIR/$btf_name"
        if [ -f "$dst_file" ]; then
            already=$((already + 1))
            continue
        fi

        # Extract to destination
        tar -xf "$btf_archive" -C "$BTF_DST_DIR/"
        count=$((count + 1))
    done
done

echo ""
echo "=== Summary ==="
echo "  New files copied:    $count"
echo "  Already existing:    $already"
echo "  Skipped (< 5.10):   $skipped"
total=$(find "$BTF_DST_DIR" -name "*.btf" -type f | wc -l)
size=$(du -sh "$BTF_DST_DIR" 2>/dev/null | cut -f1)
echo "  Total BTF files:     $total"
echo "  Total size:          $size"
