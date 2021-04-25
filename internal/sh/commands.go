package sh

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	OutGo = sh.OutCmd(mg.GoCmd())
)

