package validation

import "testing"

func TestValidateCollectResource(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		expectErr bool
	}{
		{name: "github is supported", resource: "github", expectErr: false},
		{name: "paused source is rejected", resource: "ossinsight", expectErr: true},
		{name: "unknown source is rejected", resource: "gitlab", expectErr: true},
		{name: "empty source is rejected", resource: "", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCollectResource(tt.resource)
			if tt.expectErr && err == nil {
				t.Fatalf("expected an error for resource %q", tt.resource)
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error for resource %q: %v", tt.resource, err)
			}
		})
	}
}
