package project

import (
	"github.com/ch629/mage/internal/sh"
	"strings"
	"sync"
)

var (
	pkgsInit sync.Once
	pkgs     []string
)

// Packages returns a list of package names for this module
func Packages() ([]string, error) {
	var err error
	pkgsInit.Do(func() {
		var pkgName string
		if pkgName, err = ModuleName(); err != nil {
			return
		}
		pkgLen := len(pkgName)
		var s string
		if s, err = sh.OutGo("list", "./..."); err != nil {
			return
		}
		pkgs = strings.Split(s, "\n")
		for i := range pkgs {
			pkgs[i] = "." + pkgs[i][pkgLen:]
		}
	})
	return pkgs, err
}
