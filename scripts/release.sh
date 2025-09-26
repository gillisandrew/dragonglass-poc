#!/bin/bash

# Release script for Dragonglass CLI tools
# Usage: ./scripts/release.sh dragonglass v1.0.0
#        ./scripts/release.sh dragonglass-build v1.0.0

set -euo pipefail

TOOL=${1:-}
VERSION=${2:-}

if [[ -z "$TOOL" || -z "$VERSION" ]]; then
    echo "Usage: $0 <tool> <version>"
    echo ""
    echo "Examples:"
    echo "  $0 dragonglass v1.0.0"
    echo "  $0 dragonglass-build v1.0.0"
    echo ""
    echo "Available tools:"
    echo "  - dragonglass"
    echo "  - dragonglass-build"
    exit 1
fi

# Validate tool name
if [[ "$TOOL" != "dragonglass" && "$TOOL" != "dragonglass-build" ]]; then
    echo "❌ Error: Tool must be 'dragonglass' or 'dragonglass-build'"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
    echo "❌ Error: Version must follow semver format (e.g., v1.0.0, v1.0.0-beta.1)"
    exit 1
fi

TAG="${TOOL}/${VERSION}"

echo "🏷️  Preparing to create release tag: $TAG"
echo ""

# Check if tag already exists
if git tag -l "$TAG" | grep -q "$TAG"; then
    echo "❌ Error: Tag $TAG already exists"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo "❌ Error: Working directory is not clean. Please commit your changes first."
    exit 1
fi

# Ensure we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo "⚠️  Warning: You are not on the main branch (currently on: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

# Confirm before proceeding
echo "This will:"
echo "  1. Create tag: $TAG"
echo "  2. Push tag to origin"
echo "  3. Trigger GitHub Actions workflow"
echo "  4. Create a GitHub release for $TOOL $VERSION"
echo ""
read -p "Proceed? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create and push tag
echo "🏗️  Creating tag $TAG..."
git tag "$TAG"

echo "🚀 Pushing tag to origin..."
git push origin "$TAG"

echo ""
echo "✅ Release initiated successfully!"
echo ""
echo "🔗 Check the release progress at:"
echo "   https://github.com/gillisandrew/dragonglass-poc/actions"
echo ""
echo "📦 Once complete, the release will be available at:"
echo "   https://github.com/gillisandrew/dragonglass-poc/releases/tag/$TAG"