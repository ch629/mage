package coverage

import (
	"bytes"
	"fmt"
	"github.com/ch629/mage/internal/project"
	"github.com/ch629/mage/internal/sh"
	"github.com/magefile/mage/mg"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
)

const coverageFileName = "coverage.out"

var (
	covRegex = regexp.MustCompile(`[0-9.]+%`)
	cov      []PackageTestCoverage
	covInit  sync.Once
)

type (
	// Options is the code coverage requirements
	Options struct {
		// TODO: Exclusions
		Threshold float32
	}

	// PackageTestCoverage groups up each Package with it's Coverage percentage
	PackageTestCoverage struct {
		Package     string
		Coverage    float32
		FailureText string
	}
)

// FormatCoverageReport writes the Coverage report to the io.Writer
// returns failed if any of the packages is below the threshold
// returns an error if any errors writing to the writer
func FormatCoverageReport(options Options, cov []PackageTestCoverage, w io.Writer) (belowCoverage bool, err error) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	failBuf := &bytes.Buffer{}
	for _, c := range cov {
		if c.Coverage < options.Threshold {
			belowCoverage = true
			failedStr := ""
			// Tests belowCoverage
			if len(c.FailureText) > 0 {
				failedStr = "FAILED"
				_, _ = fmt.Fprintln(failBuf)
				if _, err = fmt.Fprintln(failBuf, c.FailureText); err != nil {
					return
				}
			}

			if _, err = fmt.Fprintf(tw, "\t%v\t%05.2f%%\t%v\n", c.Package, c.Coverage, failedStr); err != nil {
				return
			}
		}
	}
	if err = tw.Flush(); err != nil {
		return
	}

	_, err = io.Copy(w, failBuf)

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
			cov[i] = PackageTestCoverage{Package: strings.TrimPrefix(pkg, "./")}
			if res, err = sh.OutGoTest(pkg, "-covermode=count"); err != nil {
				// Likely to just be a test failure
				cov[i].Coverage = 0
				cov[i].FailureText = res
				err = nil
				continue
			}
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

// Generate generates the coverage report output file
func Generate() error {
	return sh.GoTest(fmt.Sprintf("-coverprofile=%s", coverageFileName), "./...")
}

// HTMLReport creates the HTML report of an existing coverage report
func HTMLReport() error {
	mg.Deps(Generate)
	return sh.GoCover(fmt.Sprintf("-html=%s", coverageFileName))
}

// Report outputs the function level coverage of an existing report
func Report() error {
	mg.Deps(Generate)
	return sh.GoCover(fmt.Sprintf("-func=%s", coverageFileName))
}

// Cleanup deletes the coverage report file
func Cleanup() error {
	err := os.Remove(coverageFileName)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func covPercentFromLine(line string) (float32, error) {
	covPercent, err := strconv.ParseFloat(strings.TrimSuffix(covRegex.FindStringSubmatch(line)[0], "%"), 32)
	return float32(covPercent), err
}
