//go:build tools
// +build tools

// File is required to force golang to install mage despite it isn't used in the code (but in build)

package tools

import (
	_ "github.com/magefile/mage/mage"
)