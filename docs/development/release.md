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

6. **Update changelog**
   ```bash
   yarn changeset:version
   ```

7. **Update version**
   - set proper version in `package.json`

7. **Check src/plugin.json version and date**
   - `node scripts/update-versions.js`
   - Verify the `version` number and `updated` date are correct

8. **Create and push git tag**
   ```bash
   git tag x.x.x  # Use version from plugin.json
   git push && git push --tags
   ```

9. **Create GitHub release**
   - Navigate to https://github.com/HadesArchitect/GrafanaCassandraDatasource/releases
   - Click "Create a new release"
   - Select the git tag created in step 8 (e.g., `3.0.1`)
   - Set release title to match the version (e.g., `3.0.1`)
   - Copy the relevant section from `CHANGELOG.md` as the release description
   - Mark as "Latest release" if this is the newest version
   - Click "Publish release"
   - This will automatically trigger the GitHub workflow to build and attach release artifacts

10. **Submit the new zip package to the Grafana Plugins Repository**
   - Wait for the release build to finish building the package and zip file
   - https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin#publish-your-plugin (For now has to be done by @HadesArchitect)

## Notes

- The `yarn changeset:version` command will consume all pending changesets and update versions
- Always verify the version in `src/plugin.json` matches `package.json` after running changeset:version
- The git tag should match the exact version number from the plugin files
- Creating the GitHub release will automatically trigger the build workflow that creates and uploads release artifacts
- The workflow will generate `cassandra-datasource-VERSION.zip` and corresponding `.md5` files
- Verify that the workflow completes successfully and artifacts are attached to the release