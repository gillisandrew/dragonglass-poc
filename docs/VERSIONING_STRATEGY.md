# Separate Versioning Strategy for Dragonglass CLI Tools

## Current Situation
- Single Go module with both CLI tools
- Shared versioning causes unnecessary coupling
- Different audiences and release cadences

## Strategy 1: Prefixed Git Tags (Recommended)

### Advantages
✅ Simple to implement with current structure  
✅ Clear separation in releases  
✅ No breaking changes to current setup  
✅ GitHub releases are cleanly separated  

### Implementation
**Tag Format:**
- `dragonglass/v1.0.0` - Plugin manager releases
- `dragonglass-build/v1.0.0` - Build tool releases

**Distribution Channels:**
- GitHub Releases (tarballs + checksums)
- GitHub Container Registry (OCI artifacts via ORAS)

**Security Features:**
- Build provenance attestation for all binaries
- SBOM (Software Bill of Materials) generation
- Supply chain verification via Sigstore
- Attestations pushed to registry for OCI artifacts

**Usage:**
```bash
# Release plugin manager
git tag dragonglass/v1.0.0
git push origin dragonglass/v1.0.0

# Release build tool  
git tag dragonglass-build/v1.0.0
git push origin dragonglass-build/v1.0.0
```

**Installation Options:**
```bash
# Traditional download from GitHub Releases
curl -L -o dragonglass.tar.gz https://github.com/.../releases/download/...

# OCI artifact via ORAS from GHCR
oras pull ghcr.io/gillisandrew/dragonglass:v1.0.0
```

**Workflows:** Separate workflow files trigger on different tag patterns

---

## Strategy 2: Separate Go Modules

### Structure
```
dragonglass-poc/
├── go.mod (shared internal packages)
├── cmd/
│   ├── dragonglass/
│   │   ├── go.mod (dragonglass-specific deps)
│   │   └── main.go
│   └── dragonglass-build/
│       ├── go.mod (includes dagger.io/dagger)
│       └── main.go
```

### Advantages
✅ True dependency isolation  
✅ Independent Go module versioning  
✅ Smaller dependency footprints  
✅ Can use different Go versions if needed  

### Disadvantages
❌ More complex dependency management  
❌ Need to version shared internal packages  
❌ More complex build processes  

---

## Strategy 3: Hybrid Approach

### Structure
Keep current monorepo but add version metadata files:

```
cmd/
├── dragonglass/
│   ├── VERSION
│   └── main.go
└── dragonglass-build/
    ├── VERSION  
    └── main.go
```

### Implementation
- Single workflow reads version files
- Builds only if version changed since last release
- Creates separate releases with tool-specific versions

---

## Recommendation

**Start with Strategy 1 (Prefixed Tags)** because:

1. **Immediate implementation** - Works with your current structure
2. **Clean separation** - Each tool gets its own releases and documentation
3. **Independent versioning** - Can release updates to one tool without the other
4. **Familiar workflow** - Standard Git tagging, just with prefixes
5. **Easy migration** - Can evolve to other strategies later if needed

## Next Steps

1. Create separate workflow files (✅ Done)
2. Update existing release.yml to handle legacy tags or remove it
3. Test with initial releases:
   ```bash
   git tag dragonglass/v1.0.0
   git tag dragonglass-build/v1.0.0
   git push --tags
   ```

## Version Progression Examples

**Plugin Manager** (user-facing, stable):
- `dragonglass/v1.0.0` - Initial release
- `dragonglass/v1.0.1` - Bug fixes
- `dragonglass/v1.1.0` - New verification features
- `dragonglass/v2.0.0` - Major API changes

**Build Tool** (developer-facing, rapid iteration):
- `dragonglass-build/v0.1.0` - Initial alpha
- `dragonglass-build/v0.2.0` - Dagger updates
- `dragonglass-build/v0.3.0` - New build features
- `dragonglass-build/v1.0.0` - Production ready