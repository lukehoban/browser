//go:build !(js && wasm)
// +build !js !wasm

package main

// This file exists to prevent test failures when running tests
// on non-WASM platforms. The browser-wasm package can only be
// tested when compiled with GOOS=js GOARCH=wasm.
