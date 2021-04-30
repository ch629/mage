package test

import (
	"fmt"
	"github.com/ch629/mage/internal/coverage"
	"github.com/magefile/mage/mg"
	"strings"
)

// TODO: Pick these up from yml with defaults
var coverageOptions = coverage.Options{
	Threshold: 80,
}

// TODO: Do we want this to be test::report:html etc?
type Report mg.Namespace

// Check validates the test coverage
func Check() (err error) {
	cover, err := coverage.Coverage()
	if err != nil {
		return fmt.Errorf("failed to generate coverage %w", err)
	}

	sb := &strings.Builder{}
	failed, err := coverage.FormatCoverageReport(coverageOptions, cover, sb)
	if err != nil {
		return fmt.Errorf("failed to format report %w", err)
	}

	if failed {
		// TODO: This needs to change if tests are failing
		return fmt.Errorf("package test coverage too low, expected minimum %.f%%\n%s", coverageOptions.Threshold, sb)
	}

	return
}

// HTML generates a HTML coverage report and opens it in the browser
func (Report) HTML() error {
	return coverage.HTMLReport()
}

// Func reports the function test coverage
func (Report) Func() error {
	return coverage.Report()
}

// Cleanup deletes the coverage report file
func (Report) Cleanup() error {
	return coverage.Cleanup()
}
