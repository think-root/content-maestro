package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ValidateCronExpression(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) != 5 && len(fields) != 6 {
		return fmt.Errorf("invalid number of fields in cron expression")
	}

	validators := []struct {
		name     string
		min, max int
		validate func(string) bool
	}{
		{"minute", 0, 59, validateField},
		{"hour", 0, 23, validateField},
		{"day of month", 1, 31, validateField},
		{"month", 1, 12, validateField},
		{"day of week", 0, 6, validateField},
	}

	for i, field := range fields[:5] {
		if !isValidCronField(field, validators[i].min, validators[i].max, validators[i].validate) {
			return fmt.Errorf("invalid %s field: %s", validators[i].name, field)
		}
	}

	return nil
}

func isValidCronField(field string, min, max int, validate func(string) bool) bool {
	if field == "*" {
		return true
	}

	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "/") {
			if !validateStep(part, min, max) {
				return false
			}
			continue
		}

		if strings.Contains(part, "-") {
			if !validateRange(part, min, max) {
				return false
			}
			continue
		}

		if !validate(part) || !isInRange(part, min, max) {
			return false
		}
	}

	return true
}

func validateField(field string) bool {
	match, _ := regexp.MatchString(`^\d+$`, field)
	return match
}

func validateStep(field string, min, max int) bool {
	parts := strings.Split(field, "/")
	if len(parts) != 2 {
		return false
	}

	if parts[0] == "*" {
		return validateField(parts[1])
	}

	if !validateRange(parts[0], min, max) {
		return false
	}

	return validateField(parts[1])
}

func validateRange(field string, min, max int) bool {
	parts := strings.Split(field, "-")
	if len(parts) != 2 {
		return false
	}

	if !validateField(parts[0]) || !validateField(parts[1]) {
		return false
	}

	start, _ := strconv.Atoi(parts[0])
	end, _ := strconv.Atoi(parts[1])

	return start >= min && end <= max && start <= end
}

func isInRange(field string, min, max int) bool {
	num, err := strconv.Atoi(field)
	if err != nil {
		return false
	}
	return num >= min && num <= max
}
