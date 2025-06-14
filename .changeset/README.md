# Changesets Guide

This project uses [Changesets](https://github.com/changesets/changesets) to manage versioning and changelogs. Changesets help us track changes, manage semantic versioning, and generate changelogs automatically.

## What are Changesets?

Changesets are a way to manage releases in a project. They allow you to:
- Track what changes have been made
- Determine what version bump is needed (patch, minor, major)
- Generate changelogs automatically
- Coordinate releases across multiple packages (if needed)

## How to Use Changesets

### 1. Creating a Changeset

When you make changes that should be included in the next release, create a changeset:

```bash
yarn changeset
```

This will:
1. Ask you what type of change this is (patch, minor, major)
2. Ask you to describe the change
3. Create a markdown file in the `.changeset` directory

### 2. Types of Changes

- **Patch** (0.0.X): Bug fixes, small improvements that don't change the API
- **Minor** (0.X.0): New features that are backward compatible
- **Major** (X.0.0): Breaking changes that are not backward compatible

### 3. Versioning and Publishing

When you're ready to release:

```bash
# Update version numbers and generate changelog
yarn changeset:version

### 4. Checking Status

To see what changesets are pending:

```bash
yarn changeset:status
```

## Workflow Example

1. **Make your changes** to the codebase
2. **Create a changeset**: `yarn changeset`
3. **Commit both your changes and the changeset file**
4. **When ready to release**: `yarn changeset:version`
5. **Review the updated version and changelog**
6. **Commit the version changes**

## Configuration

The changeset configuration is stored in [`.changeset/config.json`](.changeset/config.json). Key settings:

- `access`: Set to "public" for open source packages
- `baseBranch`: The main branch (usually "main" or "master")
- `changelog`: Uses the default changelog generator

## Best Practices

1. **Create changesets for all user-facing changes**
2. **Write clear, descriptive changeset messages**
3. **Use appropriate semantic versioning**
4. **Review generated changelogs before publishing**
5. **Keep changesets focused on single features/fixes**

## For Grafana Plugin Development

Since this is a Grafana plugin:

- **Patch releases**: Bug fixes, performance improvements
- **Minor releases**: New features, new configuration options
- **Major releases**: Breaking changes to configuration, API changes

Remember to also update the plugin version in [`src/plugin.json`](src/plugin.json) when releasing, as Grafana uses this for plugin management.

## Useful Commands

```bash
# Create a new changeset
yarn changeset

# Check what changesets are pending
yarn changeset:status

# Update versions and generate changelog
yarn changeset:version

# Add a changeset without the interactive prompt
yarn changeset add --empty
```

## Learn More

- [Changesets Documentation](https://github.com/changesets/changesets)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
