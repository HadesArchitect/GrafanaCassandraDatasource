#!/usr/bin/env node

/**
 * Script to sync the version from package.json to src/plugin.json
 * This ensures the Grafana plugin version matches the npm package version
 */

const fs = require('fs');
const path = require('path');

const packageJsonPath = path.join(__dirname, '..', 'package.json');
const pluginJsonPath = path.join(__dirname, '..', 'src', 'plugin.json');

try {
  // Read package.json
  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
  const version = packageJson.version;

  if (!version) {
    console.error('❌ No version found in package.json');
    process.exit(1);
  }

  // Read plugin.json
  const pluginJson = JSON.parse(fs.readFileSync(pluginJsonPath, 'utf8'));
  const currentPluginVersion = pluginJson.info?.version;

  if (currentPluginVersion === version) {
    console.log(`✅ Plugin version is already up to date: ${version}`);
    process.exit(0);
  }

  // Update plugin.json version
  if (!pluginJson.info) {
    pluginJson.info = {};
  }
  pluginJson.info.version = version;

  // Write updated plugin.json
  fs.writeFileSync(pluginJsonPath, JSON.stringify(pluginJson, null, 2) + '\n');

  console.log(`✅ Updated plugin version from ${currentPluginVersion || 'undefined'} to ${version}`);
  console.log(`📝 Updated: ${pluginJsonPath}`);

} catch (error) {
  console.error('❌ Error syncing plugin version:', error.message);
  process.exit(1);
}
