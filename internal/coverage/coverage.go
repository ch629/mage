package coverage

import (
	"fmt"
	"github.com/ch629/mage/internal/project"
	"github.com/ch629/mage/internal/sh"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
)

var (
	covRegex     = regexp.MustCompile(`[0-9.]+%`)
	cov          []PackageTestCoverage
	covInit      sync.Once
)

type (
	Options struct {
		// TODO: Exclusions
		Threshold float32
	}

	PackageTestCoverage struct {
		Package  string
		Coverage float32
	}
)

// FormatCoverageReport writes the Coverage report to the io.Writer
// returns failed if any of the packages is below the threshold
// returns an error if any errors writing to the writer
func FormatCoverageReport(options Options, cov []PackageTestCoverage, w io.Writer) (failed bool, err error) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, c := range cov {
		if c.Coverage < options.Threshold {
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

// Coverage generates test Coverage percentages for each package within the module
func Coverage() ([]PackageTestCoverage, error) {
	var err error
	covInit.Do(func() {
		var pkgNames []string
		if pkgNames, err = project.Packages(); err != nil {
			return
		}
		cov = make([]PackageTestCoverage, len(pkgNames))

		for i, pkg := range pkgNames {
			var res string
			if res, err = sh.OutGo("test", pkg, "-covermode=count"); err != nil {
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
