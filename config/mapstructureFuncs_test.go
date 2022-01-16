package config

import (
	"testing"
)

func Test_parseMathExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    float64
		wantErr bool
	}{
		{
			name:    "test 1",
			expr:    "2 + 2",
			want:    4.0,
			wantErr: false,
		},
		{
			name:    "test 2",
			expr:    "5.67 / 2",
			want:    2.835,
			wantErr: false,
		},
		{
			name:    "test 3",
			expr:    "x + 2",
			want:    0,
			wantErr: true,
		},
		{
			name:    "test 4",
			expr:    "yolo!",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMathExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMathExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("err is nil but value got is also nil")
				return
			}

			if !tt.wantErr && got != nil && *got != tt.want {
				t.Errorf("parseMathExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}
