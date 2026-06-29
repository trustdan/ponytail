#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const root = path.resolve(__dirname, "..");

function binPath() {
  if (process.platform === "win32") return path.join(root, "bin", "ponytail-windows-amd64.exe");
  const goos = { darwin: "darwin", linux: "linux" }[process.platform] || process.platform;
  const goarch = { x64: "amd64", arm64: "arm64" }[process.arch] || process.arch;
  return path.join(root, "bin", `ponytail-${goos}-${goarch}`);
}

const result = spawnSync(binPath(), process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(`ponytail: no bundled binary for ${process.platform}/${process.arch}`);
  process.exit(1);
}
process.exit(result.status ?? 1);