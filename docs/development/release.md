# Release Process

This document describes the manual release process for the Grafana Cassandra Datasource plugin.

## Steps

1. **Merge PRs into main**
   - Ensure all pull requests are reviewed and merged into the main branch

2. **Pull to local**
   ```bash
   git checkout main
   git pull origin main
   ```

3. **Test with automated tests**
   ```bash
   make test
   ```

4. **Test manually with newest Grafana**
   - Test the plugin with the latest supported Grafana version

5. **Test manually with oldest supported Grafana**
   - Test the plugin with Grafana 7.4 (minimum supported version for plugin v3)

6. **Update version and changelog**
   ```bash
   yarn changeset:version
   ```

7. **Check src/plugin.json version and date**
   - Verify the version number is correct
   - Update the `updated` field with current date if needed

8. **Create and push git tag**
   ```bash
   git tag x.x.x  # Use version from plugin.json
   git push && git push --tags
   ```

## Notes

- The `yarn changeset:version` command will consume all pending changesets and update versions
- Always verify the version in `src/plugin.json` matches `package.json` after running changeset:version
- The git tag should match the exact version number from the plugin files