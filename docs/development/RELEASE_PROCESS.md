# Release Process

This document outlines the step-by-step process for releasing new versions of the Grafana Cassandra Datasource using changesets.

## Prerequisites

- Ensure you have the latest changes in the main branch
- All tests are passing
- All changes have been reviewed and merged

## Release Steps

### 1. Check Current Status

First, check what changesets are pending:

```bash
yarn changeset:status
```

This will show you:
- What packages will be bumped
- What type of version bump (patch, minor, major)
- Current version vs. new version

### 2. Review Pending Changesets

Look at the changeset files in `.changeset/` directory to review what changes will be included in the release.

### 3. Update Versions and Generate Changelog

Run the version command to:
- Update package.json version
- Generate/update CHANGELOG.md
- Remove consumed changeset files
- Update version in `src/plugin.json`

```bash
yarn changeset:version
```

### 4. Review Changes

After running the version command, review:
- The new version number in `package.json`
- The new version number in `src/plugin.json`
- The generated changelog entries
- Ensure all changes look correct

### 6. Commit Version Changes

Commit the version bump and changelog:

```bash
git add .
git commit -m "Release: Version vX.X.X"
git tag X.X.X
git push origin main
git push --tags
```

### 7. Create GitHub Release

1. Go to the GitHub repository
2. Click "Releases" → "Create a new release"
3. Tag version: `vX.X.X` (matching the tag)
4. Release title: `vX.X.X`
5. Description: Copy the relevant section from CHANGELOG.md

## Troubleshooting

### No changesets found
If you see "No changesets found", it means there are no pending changes to release. Create changesets for your changes first:

```bash
yarn changeset
```

### Version conflicts
If there are conflicts during version updates, resolve them manually and commit the resolved changes.

### Plugin version mismatch
Always ensure `src/plugin.json` version matches `package.json` version for Grafana plugin compatibility.

## Best Practices

1. **Always create changesets** for user-facing changes
2. **Review generated changelogs** before releasing
3. **Test the plugin** after version updates
4. **Keep releases focused** - don't bundle too many changes
5. **Use semantic versioning** appropriately
6. **Update plugin.json** version to match package.json

## Version Strategy

- **Patch (3.1.X)**: Bug fixes, performance improvements, documentation updates
- **Minor (3.X.0)**: New features, new configuration options, backward-compatible changes
- **Major (X.0.0)**: Breaking changes, API changes, configuration changes that require user action