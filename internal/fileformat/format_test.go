package fileformat

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
	}{
		{
			name:  "valid format",
			input: "2023-03-15T14:00:00Z.bash.bak",
			want: Format{
				Time: time.Date(2023, time.March, 15, 14, 0, 0, 0, time.UTC),
				Typ:  "bash",
				Ext:  "bak",
			},
			wantErr: false,
		},
		{
			name:    "invalid time format",
			input:   "invalid-time-format.bash.bak",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want {
					t.Errorf("Parse() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	f := New("zsh", WithExt("custom"))

	if f.Typ != "zsh" || f.Ext != "custom" {
		t.Errorf("New() failed, got = %v, want = typ = zsh, extension= custom", f)
	}
}

func TestFormatMethods(t *testing.T) {
	t.Parallel()
	f := Format{
		Time: time.Date(2023, time.March, 15, 14, 0, 0, 0, time.UTC),
		Typ:  "zsh",
		Ext:  "bak",
	}

	if !f.IsZSH() {
		t.Error("IsZSH() should return true for type zsh")
	}

	if f.IsBash() {
		t.Error("IsBash() should return false for type zsh")
	}

	expectedString := "2023-03-15T14:00:00Z.zsh.bak"
	if str := f.String(); str != expectedString {
		t.Errorf("String() got = %s, want %s", str, expectedString)
	}
}
