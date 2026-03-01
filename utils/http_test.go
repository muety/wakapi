package utils

import "testing"

func TestParseUserAgent(t *testing.T) {
	type args struct {
		ua string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "helix with wakatime-ls",
			args:    struct{ ua string }{ua: "wakatime/1.139.1 (linux-6.18.8-unknown) go1.25.5 helix/25.07.1 (74075bb5) wakatime-ls/0.2.2 helix-wakatime/0.2.2"},
			want:    "Linux",
			want1:   "helix",
			wantErr: false,
		},
		{
			name:    "stardard vscode",
			args:    struct{ ua string }{ua: "wakatime/v1.105.0 (linux-6.11.9-zen1-1-zen-unknown) go1.23.3 vscode/1.95.3 vscode-wakatime/24.8.0"},
			want:    "Linux",
			want1:   "vscode",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseUserAgent(tt.args.ua)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUserAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseUserAgent() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseUserAgent() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
