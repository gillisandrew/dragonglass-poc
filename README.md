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
      plugin-directory: '.'
    permissions:
      contents: read
      packages: write
      attestations: write
      id-token: write
```

### Advanced Usage

For plugins in subdirectories or with custom build configurations:

```yaml
name: Build and Release Plugin

on:
  push:
    tags: [ 'v*' ]

jobs:
  build:
    uses: gillisandrew/dragonglass-poc/.github/workflows/build.yml@main
    with:
      plugin-directory: 'my-plugin'
      esbuild-config: 'esbuild.config.js'
    permissions:
      contents: read
      packages: write
      attestations: write
      id-token: write
  
  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          body: |
            ## 🎉 Plugin Release
            
            **OCI Artifact:** `${{ needs.build.outputs.subject-name }}`
            **SHA256:** `${{ needs.build.outputs.subject-digest }}`
            
            ### 🔒 Security Attestations
            - **SBOM:** [${{ needs.build.outputs.sbom-attestation-url }}](${{ needs.build.outputs.sbom-attestation-url }})
            - **Provenance:** [${{ needs.build.outputs.provenance-attestation-url }}](${{ needs.build.outputs.provenance-attestation-url }})
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `plugin-directory` | Working directory containing package.json, manifest.json, and build files | ❌ | `.` |
| `esbuild-config` | Path to esbuild config file (relative to plugin-directory) | ❌ | `esbuild.config.mjs` |

## Outputs

| Output | Description |
|--------|-------------|
| `subject-name` | OCI registry name of the built artifact (e.g., `ghcr.io/owner/repo`) |
| `subject-digest` | SHA256 digest of the OCI artifact |
| `sbom-attestation-url` | URL of the SBOM attestation in GitHub |
| `provenance-attestation-url` | URL of the provenance attestation in GitHub |

## Required Files

Your plugin directory must contain these essential files:

- **`manifest.json`** - Plugin manifest with metadata (id, name, version, etc.)
- **`package.json`** - Node.js package configuration with build scripts
- **`esbuild.config.js`** - Build configuration that produces the required outputs

Your esbuild configuration must produce these three files in a `dist/` directory:

- **`main.js`** - The bundled plugin JavaScript
- **`styles.css`** - The plugin stylesheet  
- **`manifest.json`** - Copy of the plugin manifest

## Example Project Structure

```
my-plugin/
├── src/
│   ├── main.ts
│   ├── styles.scss
│   └── obsidian.d.ts
├── esbuild.config.js
├── manifest.json
├── package.json
├── tsconfig.json
└── .github/
    └── workflows/
        └── build.yml
```

### Example esbuild Configuration

```javascript
// esbuild.config.js
const esbuild = require('esbuild');
const { sassPlugin } = require('esbuild-sass-plugin');
const fs = require('fs');
const path = require('path');

async function build() {
  const isProd = process.env.NODE_ENV === 'production';
  
  console.log(`Building in ${isProd ? 'production' : 'development'} mode...`);

  // Ensure dist directory exists
  if (!fs.existsSync('dist')) {
    fs.mkdirSync('dist', { recursive: true });
  }

  // Build main.js
  await esbuild.build({
    entryPoints: ['src/main.ts'],
    bundle: true,
    outfile: 'dist/main.js',
    format: 'cjs',
    target: 'es2020',
    minify: isProd,
    sourcemap: !isProd,
    external: ['obsidian'],
  });

  // Build styles.css
  await esbuild.build({
    entryPoints: ['src/styles.scss'],
    bundle: true,
    outfile: 'dist/styles.css',
    minify: isProd,
    sourcemap: !isProd,
    plugins: [sassPlugin()],
  });

  // Copy manifest
  fs.copyFileSync('manifest.json', 'dist/manifest.json');
  
  console.log('✅ Build completed successfully!');
}

build().catch((error) => {
  console.error('❌ Build failed:', error);
  process.exit(1);
});
```

### Example manifest.json

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "minAppVersion": "0.15.0",
  "description": "Description of what my plugin does",
  "author": "Your Name",
  "authorUrl": "https://github.com/your-username/my-plugin",
  "isDesktopOnly": false
}
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