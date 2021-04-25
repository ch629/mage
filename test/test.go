package test

import (
	"fmt"
	"github.com/ch629/mage/internal/coverage"
	"strings"
)

var coverageOptions = coverage.Options{
	Threshold: 80,
}

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
		return fmt.Errorf("package test coverage too low, expected minimum %.f%%\n%s", coverageOptions.Threshold, sb)
	}

	return
}
