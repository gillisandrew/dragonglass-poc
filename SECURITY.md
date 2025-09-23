# Security and Attestation Features

This document describes the comprehensive security features built into the reusable plugin build workflow.

## Overview

The workflow implements multiple layers of security and attestation to ensure:
- **Supply chain integrity** through SLSA provenance
- **Dependency transparency** through SBOM generation
- **Artifact authenticity** through cryptographic signatures
- **Build reproducibility** through standardized environments

## SLSA (Supply-chain Levels for Software Artifacts)

### What is SLSA?
SLSA is a security framework that ensures the integrity of software artifacts and their build processes. Our workflow provides **SLSA Level 3** attestation.

### SLSA Provenance Features
- **Build environment isolation**: Runs in ephemeral GitHub-hosted runners
- **Parameterless builds**: No user-controlled parameters affect the build
- **Immutable references**: All dependencies are pinned to specific versions
- **Non-falsifiable provenance**: Cryptographically signed attestations

### Verification
```bash
# Verify SLSA provenance
gh attestation verify artifact.tar.gz --owner <owner> --type provenance

# View provenance details
gh attestation inspect artifact.tar.gz --owner <owner> --type provenance
```

## SBOM (Software Bill of Materials)

### What is an SBOM?
An SBOM is a formal record containing details and supply chain relationships of various components used in building software.

### SBOM Features
- **SPDX format**: Industry-standard format for interoperability
- **Comprehensive coverage**: All dependencies, including transitive ones
- **License tracking**: Automated license detection and reporting
- **Vulnerability correlation**: Links to known vulnerability databases

### Generated SBOM Contents
- **Package information**: Name, version, source
- **Relationship mapping**: Dependencies and their relationships
- **License data**: SPDX license identifiers
- **File checksums**: SHA-256 hashes for integrity

### Verification
```bash
# Verify SBOM attestation
gh attestation verify artifact.tar.gz --owner <owner> --type sbom

# View SBOM details
gh attestation inspect artifact.tar.gz --owner <owner> --type sbom
```

## Cryptographic Signatures

### Sigstore Integration
The workflow uses GitHub's built-in Sigstore integration for:
- **Keyless signing**: No need to manage private keys
- **Transparency logs**: All signatures logged in Rekor
- **Certificate transparency**: Build context recorded in certificates

### Signature Verification
```bash
# Using cosign (alternative method)
cosign verify-blob \
  --certificate-identity "https://github.com/<owner>/<repo>/.github/workflows/build.yml@refs/heads/main" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --signature <signature> \
  artifact.tar.gz
```

## Build Environment Security

### Container Hardening
- **Minimal base images**: Ubuntu runners with only essential tools
- **No persistent storage**: Ephemeral containers for each build
- **Network isolation**: Limited external network access
- **Resource limits**: CPU and memory constraints

### Secret Management
- **No hardcoded secrets**: All sensitive data via GitHub Secrets
- **Least privilege access**: Minimal required permissions
- **Audit logging**: All secret access logged

## Artifact Integrity

### Hash Verification
Every artifact includes multiple integrity checks:
- **SHA-256 checksums**: Primary integrity verification
- **Content validation**: Required files presence check
- **Size validation**: Artifact size tracking

### Tamper Detection
```bash
# Verify artifact hasn't been modified
echo "<expected-sha256>  artifact.tar.gz" | sha256sum -c
```

## Supply Chain Policies

### Dependency Management
- **Lock file enforcement**: Package-lock.json/yarn.lock required
- **Known vulnerability scanning**: Automated security advisories
- **License compliance**: Automated license compatibility checking

### Build Policies
- **Reproducible builds**: Deterministic build outputs
- **Source verification**: All code from tracked repositories
- **Audit trail**: Complete build history tracking

## Compliance Features

### Industry Standards
- **NIST SSDF**: Secure Software Development Framework compliance
- **SLSA**: Supply-chain Levels for Software Artifacts
- **SPDX**: Software Package Data Exchange format
- **OSSF**: Open Source Security Foundation best practices

### Regulatory Compliance
- **SOX compliance**: Audit trail and integrity controls
- **ISO 27001**: Information security management
- **GDPR**: Privacy-by-design principles

## Monitoring and Alerting

### Security Monitoring
- **Vulnerability alerts**: Automated dependency vulnerability detection
- **Anomaly detection**: Unusual build pattern identification
- **Access monitoring**: Unauthorized access attempt detection

### Audit Capabilities
```bash
# List all attestations for a repository
gh api repos/<owner>/<repo>/attestations

# Search attestations by artifact
gh api repos/<owner>/<repo>/attestations --jq '.attestations[] | select(.bundle.dsseEnvelope.payload | @base64d | fromjson | .subject[].digest.sha256 == "<artifact-sha256>")'
```

## Best Practices for Consumers

### Verification Workflow
1. **Download artifact** from trusted source
2. **Verify checksums** match expected values  
3. **Validate attestations** using GitHub CLI
4. **Check SBOM** for known vulnerabilities
5. **Verify provenance** matches expected build

### Example Verification Script
```bash
#!/bin/bash
set -euo pipefail

ARTIFACT="$1"
OWNER="$2"
EXPECTED_SHA="$3"

echo "🔍 Verifying artifact: $ARTIFACT"

# Verify checksum
ACTUAL_SHA=$(sha256sum "$ARTIFACT" | cut -d' ' -f1)
if [ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]; then
    echo "❌ Checksum mismatch!"
    exit 1
fi
echo "✅ Checksum verified"

# Verify attestations
gh attestation verify "$ARTIFACT" --owner "$OWNER"
echo "✅ Attestations verified"

# Check for vulnerabilities in SBOM
gh attestation inspect "$ARTIFACT" --owner "$OWNER" --type sbom > sbom.json
echo "✅ SBOM extracted for vulnerability scanning"

echo "🎉 Artifact verification complete!"
```

## Incident Response

### Compromise Detection
If a compromise is suspected:
1. **Revoke access**: Disable relevant GitHub tokens/secrets
2. **Audit logs**: Review all build and access logs
3. **Verify integrity**: Check all recent artifacts
4. **Update dependencies**: Refresh all dependencies

### Recovery Procedures
1. **Clean builds**: Rebuild all affected artifacts
2. **Re-attestation**: Generate new attestations
3. **Notification**: Inform downstream consumers
4. **Documentation**: Update security documentation

## Security Contacts

For security issues:
- **GitHub Security**: Use private vulnerability reporting
- **Maintainer contact**: See repository SECURITY.md
- **CVE coordination**: Through GitHub Security Advisories