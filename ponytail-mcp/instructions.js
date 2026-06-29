// Instruction selection for the Ponytail MCP server. The ruleset and default-mode
// resolution come from the ponytail binary, so every host emits identical rules.
import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function binPath() {
  if (process.platform === "win32") return path.resolve(__dirname, "../bin/ponytail-windows-amd64.exe");
  const goos = { darwin: "darwin", linux: "linux" }[process.platform] || process.platform;
  const goarch = { x64: "amd64", arm64: "arm64" }[process.arch] || process.arch;
  return path.resolve(__dirname, "../bin", `ponytail-${goos}-${goarch}`);
}
function bin(args, fallback = "") {
  try {
    return execFileSync(binPath(), args, { encoding: "utf8" });
  } catch (e) {
    return fallback;
  }
}
const getPonytailInstructions = (mode) => bin(["instructions", mode]);
const getDefaultMode = () => bin(["default-mode"], "full\n").trim();
const RUNTIME_MODES = ["off", "lite", "full", "ultra"];
const normalizeMode = (m) => {
  const n = typeof m === "string" ? m.trim().toLowerCase() : "";
  return RUNTIME_MODES.includes(n) ? n : null;
};

// The three intensities the server offers. "off" has no instructions to serve.
export const MODES = ["lite", "full", "ultra"];

// Resolve a requested mode to a runtime intensity. Unknown, empty, or "off"
// falls back to the configured default, then to "full".
// ponytail: keep the surface to these three; "off"/"review" aren't served here.
export function resolveMode(requested) {
  const asked = normalizeMode(requested);
  if (asked && asked !== "off") return asked;

  const fallback = normalizeMode(getDefaultMode());
  return fallback && fallback !== "off" ? fallback : "full";
}

export function buildInstructions(requested) {
  return getPonytailInstructions(resolveMode(requested));
}
