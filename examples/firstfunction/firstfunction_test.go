package firstfunction

import "testing"

func TestLabel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "an ordinary name", input: "demo", want: "lab-demo"},
		{name: "surrounding space is ignored", input: "  demo  ", want: "lab-demo"},
		{name: "an empty name is refused", input: "", wantErr: true},
		{name: "a name of only spaces is refused", input: "   ", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Label(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Label(%q) = %q, want an error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Label(%q) returned %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("Label(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
