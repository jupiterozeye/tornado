// Package assets contains embedded static assets for the application.
package assets

import (
	_ "embed"
)

//go:embed tornadologo.txt
var Logo string
