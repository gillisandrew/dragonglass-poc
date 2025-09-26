# Provenance Verification Guide

This guide explains how to verify the authenticity and integrity of Dragonglass CLI binaries using cryptographic provenance attestations.

## Overview

All Dragonglass CLI releases include:

- **Build Provenance Attestations** - Cryptographic proof of how binaries were built
- **SBOM (Software Bill of Materials)** - Complete dependency information
- **Checksum Verification** - File integrity validation
- **SLSA Level 3 Compliance** - Industry-standard supply chain security

## Quick Start

### 1. Install Prerequisites

**GitHub CLI:**

```bash
# macOS
brew install gh

# Ubuntu/Debian
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update && sudo apt install gh

# Windows (using winget)
winget install GitHub.cli
```

**Authenticate:**

```bash
gh auth login
```

### 2. Download and Verify

```bash
# Download release (example for macOS ARM64)
VERSION="v0.1.3"
TOOL="dragonglass"  # or "dragonglass-build"
PLATFORM="darwin-arm64"

curl -L -o ${TOOL}.tar.gz \
  "https://github.com/gillisandrew/dragonglass-poc/releases/download/${TOOL}/${VERSION}/${TOOL}-${VERSION}-${PLATFORM}.tar.gz"

# Extract
tar -xzf ${TOOL}.tar.gz

# Verify provenance
gh attestation verify ${TOOL}-${PLATFORM} --owner gillisandrew
```

## Detailed Verification Steps

### Step 1: Checksum Verification

```bash
# Download checksums
curl -L -o checksums.txt \
  "https://github.com/gillisandrew/dragonglass-poc/releases/download/${TOOL}/${VERSION}/checksums.txt"

# Verify file integrity
sha256sum -c checksums.txt
```

**Expected output:**

```
dragonglass-v0.1.3-darwin-arm64.tar.gz: OK
dragonglass-v0.1.3-darwin-amd64.tar.gz: OK
dragonglass-v0.1.3-linux-amd64.tar.gz: OK
...
```

### Step 2: Provenance Verification

```bash
# Basic verification
gh attestation verify ${TOOL}-${PLATFORM} --owner gillisandrew
```

**Expected output:**

```
Loaded digest sha256:abc123... for file://dragonglass-darwin-arm64
Loaded 1 attestation from GitHub API
✓ Verification succeeded!

sha256:abc123... was attested by:
REPO       PREDICATE_TYPE                 WORKFLOW                                   
gillisandrew/dragonglass-poc  https://slsa.dev/provenance/v1  .github/workflows/release-dragonglass.yml@refs/tags/dragonglass/v0.1.3
```

### Step 3: Detailed Provenance Analysis

```bash
# Get detailed provenance information
gh attestation verify ${TOOL}-${PLATFORM} --owner gillisandrew --format json > provenance.json

# Extract key information
echo "Built by workflow:" 
jq -r '.verificationResult.statement.predicate.buildDefinition.buildType' provenance.json

echo "From commit:"
jq -r '.verificationResult.statement.predicate.buildDefinition.resolvedDependencies[0].digest' provenance.json

echo "Build environment:"
jq -r '.verificationResult.statement.predicate.runDetails.builder.id' provenance.json
```

## OCI Artifact Verification

### Automatic Verification with ORAS

```bash
# ORAS automatically verifies attestations
oras pull ghcr.io/gillisandrew/${TOOL}:${VERSION} --verbose
```

### Manual OCI Verification

```bash
# List all attestations for the OCI artifact
oras discover ghcr.io/gillisandrew/${TOOL}:${VERSION}

# Verify OCI artifact provenance
gh attestation verify oci://ghcr.io/gillisandrew/${TOOL}:${VERSION} --owner gillisandrew
```

## Enterprise Security Practices

### Security Checklist

- [ ] **Verify checksums** - Ensure file integrity
- [ ] **Verify provenance** - Confirm authorized build process
- [ ] **Check commit SHA** - Validate source code version
- [ ] **Review SBOM** - Understand dependencies
- [ ] **Validate workflow** - Confirm expected build pipeline

### Automated Security Pipeline

```bash
#!/bin/bash
# security-verify.sh - Automated verification script

set -euo pipefail

TOOL="$1"
VERSION="$2"
PLATFORM="$3"

echo "🔍 Starting security verification for ${TOOL} ${VERSION} ${PLATFORM}"

# Download and extract
curl -L -o "${TOOL}.tar.gz" \
  "https://github.com/gillisandrew/dragonglass-poc/releases/download/${TOOL}/${VERSION}/${TOOL}-${VERSION}-${PLATFORM}.tar.gz"
tar -xzf "${TOOL}.tar.gz"

# Verify checksum
curl -L -o checksums.txt \
  "https://github.com/gillisandrew/dragonglass-poc/releases/download/${TOOL}/${VERSION}/checksums.txt"
echo "✅ Checksum verification:"
sha256sum -c checksums.txt | grep "${TOOL}-${VERSION}-${PLATFORM}.tar.gz"

# Verify provenance
echo "🔐 Provenance verification:"
gh attestation verify "${TOOL}-${PLATFORM}" --owner gillisandrew

# Extract and display key security information
gh attestation verify "${TOOL}-${PLATFORM}" --owner gillisandrew --format json > provenance.json

echo "📋 Security Details:"
echo "  Built by: $(jq -r '.verificationResult.statement.predicate.buildDefinition.buildType' provenance.json)"
echo "  Commit: $(jq -r '.verificationResult.statement.predicate.buildDefinition.resolvedDependencies[0].digest' provenance.json)"
echo "  Runner: $(jq -r '.verificationResult.statement.predicate.runDetails.builder.id' provenance.json)"

echo "✅ All security checks passed!"
```

**Usage:**

```bash
chmod +x security-verify.sh
./security-verify.sh dragonglass v0.1.3 darwin-arm64
```

## Troubleshooting

### Common Issues

**"No attestations found"**

- Ensure you're using the correct repository owner: `--owner gillisandrew`
- Check that the binary filename matches exactly
- Verify you're authenticated with GitHub: `gh auth status`

**"Verification failed"**

- The binary may have been tampered with - do not use it
- Ensure you downloaded from the official GitHub release
- Contact repository maintainers if the issue persists

**Network/Corporate Firewall Issues**

- Ensure access to:
  - `https://github.com`
  - `https://fulcio.sigstore.dev`
  - `https://rekor.sigstore.dev`
- Contact your network administrator if needed

### Getting Help

- **Repository Issues**: <https://github.com/gillisandrew/dragonglass-poc/issues>
- **GitHub CLI Help**: <https://cli.github.com/manual/>
- **SLSA Documentation**: <https://slsa.dev/>

## Security Policy

This project follows **SLSA Level 3** requirements:

- ✅ **Source integrity** - All builds from version-controlled source
- ✅ **Build integrity** - Provenance available for all artifacts  
- ✅ **Provenance available** - Cryptographic attestations provided
- ✅ **Non-falsifiable** - Provenance cannot be forged
- ✅ **Dependencies complete** - All build dependencies documented

For security concerns, please review our [Security Policy](../SECURITY.md) and report issues through appropriate channels.
