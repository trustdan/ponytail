#!/usr/bin/env node

const test = require('node:test');
const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '..');

test('npm package exposes ponytail CLI wrapper', () => {
  const pkg = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));
  assert.equal(pkg.bin.ponytail, './scripts/ponytail-bin.js');

  const result = spawnSync(process.execPath, [path.join(root, 'scripts', 'ponytail-bin.js'), 'version'], {
    encoding: 'utf8',
  });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout.trim(), pkg.version);
});