# Dragonglass Plugin Build Schema

This directory contains JSON Schema definitions for the Dragonglass plugin build system.

## build.json Schema

The `build.json` schema defines the configuration format for plugin build specifications located at:

```text
plugins/{owner}/{repo}/tags/{version}/build.json
```text

The plugin reference (tag/branch) is automatically deduced from the `{version}` part of the file path, while the specific commit hash must be explicitly provided in the build.json file.

### Schema Location

- **File**: `schemas/build.json`
- **Schema ID**: `https://raw.githubusercontent.com/gillisandrew/dragonglass-poc/refs/heads/main/schemas/build.v1.json`

### Usage

You can reference this schema in your `build.json` files:

```textjson
{
  "version": "1",
  "commit": "38d5d3fa799e1eec8f10b137eee8b2cbdc73b534"
}
```text

### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `version` | string | **yes** | Schema version (must be "1") |
| `commit` | string | **yes** | Specific commit hash to build from (40-character hex string) |
| `pluginDirectory` | string | optional | Directory within the repository containing the plugin source |
| `buildDirectory` | string | optional | Directory where `npm run build` outputs artifacts |
| `outputDirectory` | string | optional | Directory for final built artifacts (default: "dist") |

### Examples

#### Build from specific commit

```textjson
{
  "version": "1",
  "commit": "38d5d3fa799e1eec8f10b137eee8b2cbdc73b534"
}
```text

#### Build from tag with custom directories

```textjson
{
  "version": "1",
  "commit": "def456abc789123456789012345678901234abcd",
  "pluginDirectory": "src",
  "buildDirectory": "dist",
  "outputDirectory": "build"
}
```text

#### Minimal configuration

```textjson
{
  "version": "1",
  "commit": "abc123def456789012345678901234567890abcde",
  "outputDirectory": "dist"
}
```text

### Validation

The schema enforces:

- **Version field is required** and must be "1"
- **Commit field is required** and must be exactly 40 hexadecimal characters
- No additional properties are allowed
- Plugin reference (tag/branch) is automatically deduced from the file path structure

### IDE Integration

Most modern IDEs and editors will automatically provide validation, autocompletion, and documentation when the `$schema` property is included in your `build.json` files.
