package xtuis

import (
	_ "embed"
)

//go:embed version
var version string

// Version 版本信息
func Version() string {
	return version
}
