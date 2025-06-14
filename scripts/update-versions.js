const fs = require('fs');
const path = require('path');

// Read package.json version
const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
const version = packageJson.version;

// Update plugin.json version
const pluginJsonPath = path.join('src', 'plugin.json');
const pluginJson = JSON.parse(fs.readFileSync(pluginJsonPath, 'utf8'));
pluginJson.info.version = version;
fs.writeFileSync(pluginJsonPath, JSON.stringify(pluginJson, null, 2) + '\n');

console.log(`Updated version to ${version} in src/plugin.json`);
