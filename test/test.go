//+build mage

package test

import (
	"bufio"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
)

var (
	runGo = sh.RunCmd(mg.GoCmd())
	outGo = sh.OutCmd(mg.GoCmd())

	pkgs     []string
	pkgsInit sync.Once

	covRegex     = regexp.MustCompile(`[0-9.]+%`)
	cov          []PackageTestCoverage
	covInit      sync.Once
	covThreshold = float32(80)

	modInit sync.Once
	modName string
)

const (
	modulePrefix = "module "
)

type (
	Test mg.Namespace

	PackageTestCoverage struct {
		Package  string
		Coverage float32
	}
)

// Check validates the test coverage
func (Test) Check() (err error) {
	// TODO: Exclude packages & custom coverage threshold
	cover, err := coverage()
	if err != nil {
		return fmt.Errorf("failed to generate coverage %w", err)
	}

	sb := &strings.Builder{}
	failed, err := formatCoverageReport(cover, sb)
	if err != nil {
		return fmt.Errorf("failed to format report %w", err)
	}

	if failed {
		return fmt.Errorf("package test coverage too low, expected minimum %.f%%\n%s", covThreshold, sb)
	}

	return
}

// formatCoverageReport writes the coverage report to the io.Writer
// returns failed if any of the packages is below the threshold
// returns an error if any errors writing to the writer
func formatCoverageReport(cov []PackageTestCoverage, w io.Writer) (failed bool, err error) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, c := range cov {
		if c.Coverage < covThreshold {
			failed = true
			if _, err = fmt.Fprintf(tw, "\t%v\t%05.2f%%\n", c.Package, c.Coverage); err != nil {
				return
			}
		}
	}
	if err = tw.Flush(); err != nil {
		return
	}

	return
}

// packages returns a list of package names for this module
func packages() ([]string, error) {
	var err error
	pkgsInit.Do(func() {
		var pkgName string
		if pkgName, err = moduleName(); err != nil {
			return
		}
		pkgLen := len(pkgName)
		var s string
		if s, err = outGo("list", "./..."); err != nil {
			return
		}
		pkgs = strings.Split(s, "\n")
		for i := range pkgs {
			pkgs[i] = "." + pkgs[i][pkgLen:]
		}
	})
	return pkgs, err
}

// coverage generates test coverage percentages for each package within the module
func coverage() ([]PackageTestCoverage, error) {
	var err error
	covInit.Do(func() {
		var pkgNames []string
		if pkgNames, err = packages(); err != nil {
			return
		}
		cov = make([]PackageTestCoverage, len(pkgNames))

		for i, pkg := range pkgNames {
			var res string
			if res, err = outGo("test", pkg, "-covermode=count"); err != nil {
				return
			}
			cov[i] = PackageTestCoverage{Package: strings.TrimPrefix(pkg, "./")}
			if strings.Contains(res, "no test files") {
				cov[i].Coverage = 0
			} else {
				var covPercent float32
				if covPercent, err = covPercentFromLine(res); err != nil {
					return
				}
				cov[i].Coverage = covPercent
			}
		}
	})

	return cov, err
}

func covPercentFromLine(line string) (float32, error) {
	covPercent, err := strconv.ParseFloat(strings.TrimSuffix(covRegex.FindStringSubmatch(line)[0], "%"), 32)
	return float32(covPercent), err
}

// moduleName gets the name of the module based on go.mod
func moduleName() (string, error) {
	var err error
	modInit.Do(func() {
		var modFile *os.File
		if modFile, err = os.Open("go.mod"); err != nil {
			return
		}
		defer modFile.Close()

		scanner := bufio.NewScanner(modFile)
		scanner.Scan()
		// TODO: This assumes that the module is always on the first line
		modName = strings.TrimPrefix(scanner.Text(), modulePrefix)
	})
	return modName, err
}
