package sh

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	goCmd     = mg.GoCmd()
	OutGo     = sh.OutCmd(goCmd)
	OutGoTest = sh.OutCmd(goCmd, "test")
	GoTest    = sh.RunCmd(goCmd, "test")
	GoCover   = sh.RunCmd(goCmd, "tool", "cover")
)
