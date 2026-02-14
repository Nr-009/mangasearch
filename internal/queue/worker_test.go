package queue
import "testing"

func TestParsePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSeries  string
		wantChapter string
		wantPage    string
		wantErr     bool
	}{
		{
			name:        "happy path",
			input:       "/manga/Berserk/Chapter_057/014.jpg",
			wantSeries:  "Berserk",
			wantChapter: "Chapter_057",
			wantPage:    "014.jpg",
			wantErr:     false,
		},
		{
			name:        "deep path still works",
			input:       "/Users/nicolas/Downloads/manga/OnePiece/Vol_01/001.png",
			wantSeries:  "OnePiece",
			wantChapter: "Vol_01",
			wantPage:    "001.png",
			wantErr:     false,
		},
		{
			name:    "too short — only filename",
			input:   "/014.jpg",
			wantErr: true,
		},
		{
			name:    "too short — one level",
			input:   "Berserk/014.jpg",
			wantErr: true,
		},
		{
			name:        "no leading slash",
			input:       "Berserk/Chapter_057/014.jpg",
			wantSeries:  "Berserk",
			wantChapter: "Chapter_057",
			wantPage:    "014.jpg",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			series, chapter, page, err := parsePath(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if series != tt.wantSeries {
				t.Errorf("series: got %q, want %q", series, tt.wantSeries)
			}
			if chapter != tt.wantChapter {
				t.Errorf("chapter: got %q, want %q", chapter, tt.wantChapter)
			}
			if page != tt.wantPage {
				t.Errorf("page: got %q, want %q", page, tt.wantPage)
			}
		})
	}
}

