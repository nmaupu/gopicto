package config

import "testing"

func TestPDF_GetNbPictoPages(t *testing.T) {
	type fields struct {
		Page       Page
		Text       Text
		ImageWords []ImageWord
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "test 1",
			fields: fields{
				Page: Page{
					Cols:  2,
					Lines: 2,
				},
				ImageWords: make([]ImageWord, 2),
			},
			want: 1,
		},
		{
			name: "test 2",
			fields: fields{
				Page: Page{
					Cols:  2,
					Lines: 2,
				},
				ImageWords: make([]ImageWord, 0),
			},
			want: 0,
		},
		{
			name: "test 3",
			fields: fields{
				Page: Page{
					Cols:  2,
					Lines: 2,
				},
				ImageWords: make([]ImageWord, 4),
			},
			want: 1,
		},
		{
			name: "test 4",
			fields: fields{
				Page: Page{
					Cols:  2,
					Lines: 2,
				},
				ImageWords: make([]ImageWord, 5),
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PDF{
				Page:       tt.fields.Page,
				Text:       tt.fields.Text,
				ImageWords: tt.fields.ImageWords,
			}
			if got := p.GetNbPictoPages(); got != tt.want {
				t.Errorf("GetNbPictoPages() = %v, want %v", got, tt.want)
			}
		})
	}
}
