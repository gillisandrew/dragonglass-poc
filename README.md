# Dragonglass

**A secure plugin manager for Obsidian with supply chain verification**

⚠️ **Experimental Project**: This is a proof-of-concept for secure Obsidian plugin distribution. Use with caution in production environments.

## Overview

Dragonglass is a CLI tool that provides secure plugin management for Obsidian by verifying provenance attestations and Software Bill of Materials (SBOM). It addresses supply chain security concerns by establishing a verified plugin distribution channel using [SLSA](https://slsa.dev/) attestations, [Sigstore](https://www.sigstore.dev/) signatures, and [GitHub Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds).

## Why Dragonglass?

The current Obsidian plugin ecosystem lacks:
- **Supply chain security verification** - No way to verify plugins haven't been tampered with
- **Continuous vulnerability scanning** - No automated security assessment of dependencies  
- **Provenance attestation** - No proof of where and how plugins were built
- **Software Bill of Materials transparency** - No visibility into plugin dependencies

Dragonglass creates a curated ecosystem where plugins must be built through verified workflows with cryptographic attestations, ensuring users can trust the plugins they install.

## Features

### ✅ Current Capabilities

- 🔒 **Provenance Verification** - Validates plugins were built through authorized workflows using [SLSA](https://slsa.dev/) attestations
- 🛡️ **Supply Chain Security** - Verifies cryptographic signatures using [Sigstore](https://www.sigstore.dev/) and [GitHub Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)
- 📋 **SBOM Validation** - Checks for the existence of Software Bill of Materials (SBOM) attestations
- 🏗️ **Workflow Verification** - Ensures plugins were built using the expected trusted workflow
- 🏪 **Curated Ecosystem** - Only plugins built through the verified workflow are supported
- 🔑 **GitHub Integration** - Seamless authentication with GitHub App and OAuth device flow

### 🚧 Planned Features

- 🔍 **Vulnerability Scanning** - Cross-reference SBOM contents against vulnerability databases (CVE, etc.)
- 📊 **Dependency Analysis** - Deep inspection of SBOM contents for security insights
- 🚨 **Security Alerts** - Notifications when vulnerabilities are discovered in installed plugins
- 🔄 **Automatic Updates** - Secure plugin update mechanism with attestation re-verification

## Installation

### From GitHub Releases

Download the latest binary from the [releases page](https://github.com/gillisandrew/dragonglass-poc/releases) (coming soon).

### Using ORAS

Install using [ORAS](https://oras.land/) to pull directly from the OCI registry:

```bash
# Pull the latest version
oras pull ghcr.io/gillisandrew/dragonglass-poc:latest

# Install to your PATH
chmod +x dragonglass
sudo mv dragonglass /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/gillisandrew/dragonglass-poc.git
cd dragonglass-poc
make build
```

## Quick Start

1. **Authenticate with GitHub**:
   ```bash
   dragonglass auth
   ```

2. **Navigate to your Obsidian vault and install a plugin**:
   ```bash
   cd /path/to/your/vault
   dragonglass install callumalpass/tasknotes@3.23.4
   ```

3. **List installed plugins**:
   ```bash
   dragonglass list
   ```

4. **Verify plugin integrity**:
   ```bash
   dragonglass verify
   ```

## Commands

### `dragonglass auth`
Authenticate with GitHub using OAuth device flow. Credentials are securely stored in your system keychain.

### `dragonglass install <plugin>[@version]`
Install a verified plugin from the curated registry. Downloads the plugin, verifies all attestations, and installs to your Obsidian vault.

### `dragonglass list`
Show all plugins managed by Dragonglass in the current vault, including version and verification status.

### `dragonglass verify`
Re-verify all installed plugins against their attestations to ensure integrity.

## Supported Plugins

Currently supported plugins in the curated ecosystem:

- **[TaskNotes](https://github.com/callumalpass/tasknotes)** by callumalpass - Task and project management
- **[Templater](https://github.com/SilentVoid13/Templater)** by SilentVoid13 - Template engine for dynamic content

More plugins will be added as they adopt the verified build workflow. See the `plugins/` directory for the complete list.

## For Plugin Developers

### Using the Build Workflow

To have your plugin included in the Dragonglass ecosystem, use the reusable build workflow in your repository:

```yaml
name: Build and Attest Plugin

on:
  push:
    tags: ['v*']

jobs:
  build:
    uses: gillisandrew/dragonglass-poc/.github/workflows/build.yml@main
    with:
      plugin-directory: '.'
    permissions:
      contents: read
      packages: write
      attestations: write
      id-token: write
```

### Requirements

Your plugin repository must have:
- `manifest.json` - Obsidian plugin manifest (present in plugin directory)
- `package.json` - Node.js build configuration with build scripts
- Build configuration (e.g., `esbuild.config.mjs`) that supports production builds

The command `NODE_ENV=production npm run build` must produce:
- `main.js` - Bundled plugin code
- `styles.css` - Plugin styles

The build workflow will automatically copy `manifest.json` to the final artifact.

## Security Architecture

Dragonglass implements multiple layers of security verification:

### Currently Implemented
1. **Build Provenance Verification** - Validates [SLSA](https://slsa.dev/) Level 3 attestations to prove plugins were built by the expected authorized workflow
2. **Cryptographic Signature Verification** - Verifies all artifacts are properly signed using [Sigstore](https://www.sigstore.dev/) and [GitHub's attestation framework](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)
3. **SBOM Attestation Validation** - Confirms the presence of SPDX-format Software Bill of Materials attestations
4. **Workflow Identity Verification** - Ensures plugins were built by the trusted workflow identity
5. **OCI Distribution** - Secure plugin distribution through [OCI-compliant](https://opencontainers.org/) registries using [ORAS](https://oras.land/)

### Planned Security Enhancements
- **SBOM Content Analysis** - Deep inspection of dependency lists in SBOM attestations
- **Vulnerability Database Integration** - Cross-reference SBOM dependencies against CVE databases
- **Automated Security Scanning** - Continuous monitoring for newly discovered vulnerabilities

## Configuration

Dragonglass stores configuration in `.obsidian/dragonglass.json` within each vault:

```json
{
  "version": "1.0",
  "trustedBuilder": "https://github.com/gillisandrew/dragonglass-poc/.github/workflows/build.yml@refs/heads/main",
  "annotationNamespace": "md.obsidian.plugin.v0",
  "plugins": []
}
```

## Roadmap

- [ ] **Pre-built Binaries** - GitHub releases with signed binaries for all platforms
- [ ] **Package Manager Integration** - Homebrew, apt, and other package managers
- [ ] **Plugin Discovery** - Browse and search available verified plugins
- [ ] **Vulnerability Alerts** - Notifications when installed plugins have security issues
- [ ] **Automatic Updates** - Secure plugin update mechanism with attestation verification
- [ ] **Community Curation** - Process for plugin developers to request inclusion

## Related Projects

- **[SLSA](https://slsa.dev/)** - Supply-chain Levels for Software Artifacts framework
- **[Sigstore](https://www.sigstore.dev/)** - Software signing and verification infrastructure  
- **[GitHub Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)** - GitHub's artifact attestation system
- **[ORAS](https://oras.land/)** - OCI Registry As Storage for distributing artifacts
- **[in-toto](https://in-toto.io/)** - Framework for securing software supply chains
- **[SPDX](https://spdx.dev/)** - Software Package Data Exchange for SBOMs

## Contributing

This is an experimental project exploring secure plugin distribution. Contributions, feedback, and security reviews are welcome!

## License

MIT - See [LICENSE](LICENSE) for details.

**Note**: Dragonglass is licensed under MIT, but each curated plugin retains its own license terms. Please check individual plugin repositories for their specific licensing requirements.