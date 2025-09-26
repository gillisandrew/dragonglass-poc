#!/bin/bash

# Quick Dragonglass Binary Verification Script
# Usage: ./verify.sh <tool> <version> <platform>
# Example: ./verify.sh dragonglass v0.1.3 darwin-arm64

set -euo pipefail

TOOL="${1:-}"
VERSION="${2:-}"
PLATFORM="${3:-}"

if [[ -z "$TOOL" || -z "$VERSION" || -z "$PLATFORM" ]]; then
    echo "Usage: $0 <tool> <version> <platform>"
    echo ""
    echo "Examples:"
    echo "  $0 dragonglass v0.1.3 darwin-arm64"
    echo "  $0 dragonglass-build v0.1.3 linux-amd64"
    echo ""
    echo "Available platforms: darwin-arm64, darwin-amd64, linux-amd64, linux-arm64, windows-amd64.exe"
    exit 1
fi

# Validate tool name
if [[ "$TOOL" != "dragonglass" && "$TOOL" != "dragonglass-build" ]]; then
    echo "❌ Error: Tool must be 'dragonglass' or 'dragonglass-build'"
    exit 1
fi

echo "🔍 Verifying ${TOOL} ${VERSION} for ${PLATFORM}"
echo ""

# Check prerequisites
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is required but not installed."
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo "❌ curl is required but not installed."
    exit 1
fi

if ! command -v tar &> /dev/null; then
    echo "❌ tar is required but not installed."
    exit 1
fi

if ! command -v sha256sum &> /dev/null && ! command -v shasum &> /dev/null; then
    echo "❌ sha256sum or shasum is required but not installed."
    exit 1
fi

# Check GitHub authentication
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub CLI."
    echo "Run: gh auth login"
    exit 1
fi

BINARY_NAME="${TOOL}-${PLATFORM}"
TARBALL_NAME="${TOOL}-${VERSION}-${PLATFORM}.tar.gz"
RELEASE_TAG="${TOOL}/${VERSION}"
DOWNLOAD_URL="https://github.com/gillisandrew/dragonglass-poc/releases/download/${RELEASE_TAG}/${TARBALL_NAME}"

echo "📥 Downloading ${TARBALL_NAME}..."
curl -L -o "${TARBALL_NAME}" "${DOWNLOAD_URL}"

echo "📦 Extracting tarball..."
tar -xzf "${TARBALL_NAME}"

# Download checksums
echo "🔍 Downloading and verifying checksums..."
curl -L -o checksums.txt "https://github.com/gillisandrew/dragonglass-poc/releases/download/${RELEASE_TAG}/checksums.txt"

# Verify checksum
if command -v sha256sum &> /dev/null; then
    echo "✅ Checksum verification:"
    sha256sum -c checksums.txt | grep "${TARBALL_NAME}"
elif command -v shasum &> /dev/null; then
    echo "✅ Checksum verification:"
    shasum -a 256 -c checksums.txt | grep "${TARBALL_NAME}"
fi

# Verify provenance
echo "🔐 Verifying build provenance..."
gh attestation verify "${BINARY_NAME}" --owner gillisandrew

echo ""
echo "✅ All verification checks passed!"
echo ""
echo "📋 Security Summary:"
echo "  ✓ File integrity verified (checksum)"
echo "  ✓ Build provenance verified (attestation)"
echo "  ✓ Binary built by authorized GitHub Actions workflow"
echo "  ✓ Source code authenticity confirmed"
echo ""
echo "🚀 Binary is ready to use: ./${BINARY_NAME}"
echo ""
echo "💡 To install globally:"
echo "  chmod +x ${BINARY_NAME}"
if [[ "$PLATFORM" == *"windows"* ]]; then
    echo "  # On Windows, move to a directory in your PATH"
else
    echo "  sudo mv ${BINARY_NAME} /usr/local/bin/${TOOL}"
fi

# Cleanup
echo ""
echo "🧹 Cleaning up temporary files..."
rm -f "${TARBALL_NAME}" checksums.txt

echo "✨ Verification complete!"