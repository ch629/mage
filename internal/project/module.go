package project

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

var (
	modInit sync.Once
	modName string
)

const (
	modulePrefix = "module "
)

// ModuleName gets the name of the module based on go.mod
func ModuleName() (string, error) {
	var err error
	modInit.Do(func() {
		var modFile *os.File
		if modFile, err = os.Open("go.mod"); err != nil {
			return
		}
		defer modFile.Close()

		scanner := bufio.NewScanner(modFile)
		scanner.Scan()
		modName = strings.TrimPrefix(scanner.Text(), modulePrefix)
	})
	return modName, err
}
