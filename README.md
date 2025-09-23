# Dragonglass POC - Reusable Plugin Build Workflow

This repository provides a reusable GitHub Actions workflow for building and attesting esbuild plugins with comprehensive security attestations.

## Features

- 🔧 **Automated esbuild bundling** with configurable build commands
- 📦 **Standardized artifact packaging** (tarball with main.js, styles.css, manifest.json)
- 🔒 **SBOM generation and attestation** using Anchore SBOM action
- 🛡️ **SLSA provenance attestation** for supply chain security
- 🎯 **Flexible configuration** supporting npm, yarn, and pnpm
- ♻️ **Reusable workflow** for multiple plugin projects

## Usage

### Basic Usage

Create a workflow in your plugin repository that calls this reusable workflow:

```yaml
name: Build Plugin

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    uses: gillisandrew/dragonglass-poc/.github/workflows/build.yml@main
    with:
      plugin-name: 'my-awesome-plugin'
    permissions:
      contents: read
      packages: write
      attestations: write
      id-token: write
```

### Advanced Usage

```yaml
name: Build and Release Plugin

on:
  push:
    tags: [ 'v*' ]

jobs:
  build:
    uses: gillisandrew/dragonglass-poc/.github/workflows/build.yml@main
    with:
      plugin-name: 'my-plugin'
      node-version: '18'
      build-command: 'npm run build:prod'
      source-directory: 'build'
      package-manager: 'yarn'
    permissions:
      contents: read
      packages: write
      attestations: write
      id-token: write
  
  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download artifact
        uses: actions/download-artifact@v4
        with:
          name: ${{ needs.build.outputs.artifact-name }}
      
      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          files: ${{ needs.build.outputs.artifact-name }}
          body: |
            ## 🎉 Plugin Release
            
            **Artifact:** `${{ needs.build.outputs.artifact-name }}`
            **SHA256:** `${{ needs.build.outputs.artifact-digest }}`
            
            ### 🔒 Security Attestations
            - **SBOM:** `${{ needs.build.outputs.sbom-attestation-id }}`
            - **Provenance:** `${{ needs.build.outputs.provenance-attestation-id }}`
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `plugin-name` | Name of the plugin (used for tarball naming) | ✅ | - |
| `node-version` | Node.js version to use | ❌ | `20` |
| `build-command` | Command to run esbuild | ❌ | `npm run build` |
| `source-directory` | Directory containing built files | ❌ | `dist` |
| `package-manager` | Package manager (npm, yarn, pnpm) | ❌ | `npm` |

## Outputs

| Output | Description |
|--------|-------------|
| `artifact-name` | Name of the built artifact |
| `artifact-digest` | SHA256 digest of the artifact |
| `sbom-attestation-id` | ID of the SBOM attestation |
| `provenance-attestation-id` | ID of the provenance attestation |

## Required Files

Your esbuild configuration must produce these three files in the specified source directory:

- **`main.js`** - The bundled plugin JavaScript
- **`styles.css`** - The plugin stylesheet
- **`manifest.json`** - The plugin manifest

## Example Project Structure

```
my-plugin/
├── src/
│   ├── main.ts
│   ├── styles.scss
│   └── manifest.json
├── esbuild.config.js
├── package.json
└── .github/
    └── workflows/
        └── build.yml
```

### Example esbuild Configuration

```javascript
// esbuild.config.js
const esbuild = require('esbuild');
const fs = require('fs');
const path = require('path');

async function build() {
  // Build main.js
  await esbuild.build({
    entryPoints: ['src/main.ts'],
    bundle: true,
    outfile: 'dist/main.js',
    format: 'cjs',
    target: 'es2020',
    minify: true,
  });

  // Build styles.css
  await esbuild.build({
    entryPoints: ['src/styles.scss'],
    bundle: true,
    outfile: 'dist/styles.css',
    minify: true,
  });

  // Copy manifest
  fs.copyFileSync('src/manifest.json', 'dist/manifest.json');
}

build().catch(() => process.exit(1));
```

### Example package.json

```json
{
  "name": "my-plugin",
  "scripts": {
    "build": "node esbuild.config.js",
    "build:prod": "NODE_ENV=production node esbuild.config.js"
  },
  "devDependencies": {
    "esbuild": "^0.19.0",
    "typescript": "^5.0.0"
  }
}
```

## Security Features

### SBOM (Software Bill of Materials)
- Generated using Anchore SBOM action
- Provides comprehensive dependency tracking
- Attested and signed using GitHub's attestation framework

### SLSA Provenance
- Level 3 SLSA provenance attestation
- Verifiable build integrity
- Supply chain security guarantees

### Attestation Verification

You can verify the attestations using the GitHub CLI:

```bash
# Verify SBOM attestation
gh attestation verify artifact.tar.gz --owner gillisandrew

# Verify provenance attestation  
gh attestation verify artifact.tar.gz --owner gillisandrew --type provenance
```

## Permissions Required

The calling workflow must include these permissions:

```yaml
permissions:
  contents: read      # Read repository contents
  packages: write     # Upload artifacts
  attestations: write # Create attestations
  id-token: write     # OIDC token for signing
```

## License

MIT