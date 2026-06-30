module github.com/DietrichGebert/ponytail

go 1.26

// Pin the exact toolchain so dev and CI build byte-identical binaries (bin/ is
// committed and CI diffs a fresh rebuild). Bump this in lockstep with a rebuild.
toolchain go1.26.4
