import assert from "node:assert/strict";
import { existsSync, mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
  parsePonytailCommand,
  readDefaultMode,
  resolveSessionMode,
  writeDefaultMode,
} from "../index.js";

test("parsePonytailCommand falls back to full when invoked bare and default is off", () => {
  assert.deepEqual(parsePonytailCommand("", "off"), { type: "set-mode", mode: "full" });
});

test("parsePonytailCommand parses modes, status, and default subcommand", () => {
  assert.deepEqual(parsePonytailCommand("ultra", "full"), { type: "set-mode", mode: "ultra" });
  assert.deepEqual(parsePonytailCommand("status", "full"), { type: "status" });
  assert.deepEqual(parsePonytailCommand("default lite", "full"), { type: "set-default", mode: "lite" });
});

test("resolveSessionMode prefers latest persisted session mode", () => {
  const entries = [
    { type: "custom", customType: "ponytail-mode", data: { mode: "lite" } },
    { type: "custom", customType: "ponytail-mode", data: { mode: "ultra" } },
  ];

  assert.equal(resolveSessionMode(entries, "full"), "ultra");
});

test("resolveSessionMode returns fallback when entries is not an array", () => {
  assert.equal(resolveSessionMode(null, "ultra"), "ultra");
  assert.equal(resolveSessionMode(undefined, "lite"), "lite");
  assert.equal(resolveSessionMode({}, "full"), "full");
  assert.equal(resolveSessionMode("not an array"), "full"); // DEFAULT_MODE fallback
});

test("readDefaultMode and writeDefaultMode use XDG config path", () => {
  const tempDir = mkdtempSync(join(tmpdir(), "ponytail-config-"));
  const previousXdg = process.env.XDG_CONFIG_HOME;
  const previousDefault = process.env.PONYTAIL_DEFAULT_MODE;
  const configPath = join(tempDir, "ponytail", "config.json");
  process.env.XDG_CONFIG_HOME = tempDir;
  delete process.env.PONYTAIL_DEFAULT_MODE;

  try {
    assert.equal(readDefaultMode(), "full");
    assert.equal(writeDefaultMode("ultra"), "ultra");
    assert.equal(readDefaultMode(), "ultra");
    assert.ok(existsSync(configPath));
    assert.deepEqual(JSON.parse(readFileSync(configPath, "utf8")), { defaultMode: "ultra" });
  } finally {
    if (previousXdg === undefined) delete process.env.XDG_CONFIG_HOME;
    else process.env.XDG_CONFIG_HOME = previousXdg;
    if (previousDefault === undefined) delete process.env.PONYTAIL_DEFAULT_MODE;
    else process.env.PONYTAIL_DEFAULT_MODE = previousDefault;
    rmSync(tempDir, { recursive: true, force: true });
  }
});