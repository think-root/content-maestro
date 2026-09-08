package validation

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var validCollectResources = []string{"github"}

// pausedCollectResources are sources whose upstream ranking is unavailable, so a
// collect run against them returns nothing. OssInsight paused its star-based
// rankings on 2026-03-01; move it back to validCollectResources once the metric
// returns.
var pausedCollectResources = []string{"ossinsight"}

// ValidateCollectResource reports whether resource can be used for a collect run.
// An empty value is left to the caller to default.
func ValidateCollectResource(resource string) error {
	if slices.Contains(pausedCollectResources, resource) {
		return fmt.Errorf("resource %q is paused upstream and cannot be scheduled", resource)
	}

	if !slices.Contains(validCollectResources, resource) {
		return errors.New("invalid resource, allowed values: " + strings.Join(validCollectResources, ", "))
	}

	return nil
}
