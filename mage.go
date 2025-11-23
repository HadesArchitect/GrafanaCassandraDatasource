//go:build ignore
// +build ignore

// File is used to build the plugin without installing mage (see Makefile be-build)

package main

import (
    "os"
    "github.com/magefile/mage/mage"
)

func main() { os.Exit(mage.Main()) }
