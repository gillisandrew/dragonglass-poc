# Dragonglass Plugin Build Schema

This directory contains JSON Schema definitions for the Dragonglass plugin build system.

## build.json Schema

The `build.json` schema defines the configuration format for plugin build specifications located at:
```
plugins/{owner}/{repo}/tags/{version}/build.json
```

### Schema Location

- **File**: `schemas/build.json`
- **Schema ID**: `https://github.com/gillisandrew/dragonglass-poc/schemas/build.json`

### Usage

You can reference this schema in your `build.json` files:

```json
{
  "$schema": "../../../../../schemas/build.json",
  "commit": "38d5d3fa799e1eec8f10b137eee8b2cbdc73b534"
}
```

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `commit` | string | conditional* | Specific commit hash to build from (40-character hex string) |
| `pluginRef` | string | conditional* | Git reference (branch or tag) to build from |
| `pluginDirectory` | string | optional | Directory within the repository containing the plugin source |
| `buildDirectory` | string | optional | Directory where `npm run build` outputs artifacts |
| `outputDirectory` | string | optional | Directory for final built artifacts (default: "dist") |

*Either `commit` or `pluginRef` must be specified, or both can be omitted to use the version from the file path.

### Examples

#### Build from specific commit:
```json
{
  "commit": "38d5d3fa799e1eec8f10b137eee8b2cbdc73b534"
}
```

#### Build from tag with custom directories:
```json
{
  "pluginRef": "v2.15.2",
  "pluginDirectory": "src",
  "buildDirectory": "dist",
  "outputDirectory": "build"
}
```

#### Minimal configuration (uses version from path):
```json
{
  "outputDirectory": "dist"
}
```

### Validation

The schema enforces:
- Commit hashes must be exactly 40 hexadecimal characters
- No additional properties are allowed
- At least one of `commit` or `pluginRef` should be specified (or neither to use path-inferred version)

### IDE Integration

Most modern IDEs and editors will automatically provide validation, autocompletion, and documentation when the `$schema` property is included in your `build.json` files.