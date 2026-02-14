package watcher

import (
	"context"
	"sort"
	"testing"
	"time"
)

type mockDB struct {
	snapshots map[string]time.Time
}

func (m *mockDB) LoadSnapshots(ctx context.Context) (map[string]time.Time, error) {
	return m.snapshots, nil
}


func TestIsImageFile(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"014.jpg", true},
		{"014.jpeg", true},
		{"014.png", true},
		{"014.JPG", true},
		{"014.PNG", true},
		{"readme.txt", false},
		{"archive.zip", false},
		{"chapter.pdf", false},
		{"data.cbz", false},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isImageFile(tt.input)
			if got != tt.want {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- Compare ---

func TestCompareWithoutScan(t *testing.T) {
	now := time.Now()
	old := now.Add(-1 * time.Hour)

	tests := []struct {
		name         string
		filesFound   map[string]time.Time
		snapshots    map[string]time.Time
		wantToIndex  []string
		wantToDelete []string
	}{
		{
			name: "new file — not in DB",
			filesFound: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": now,
			},
			snapshots:    map[string]time.Time{},
			wantToIndex:  []string{"/manga/Berserk/Chapter_057/014.jpg"},
			wantToDelete: []string{},
		},
		{
			name:       "deleted file — in DB but not on disk",
			filesFound: map[string]time.Time{},
			snapshots: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": old,
			},
			wantToIndex:  []string{},
			wantToDelete: []string{"/manga/Berserk/Chapter_057/014.jpg"},
		},
		{
			name: "modified file — newer timestamp on disk",
			filesFound: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": now,
			},
			snapshots: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": old,
			},
			wantToIndex:  []string{"/manga/Berserk/Chapter_057/014.jpg"},
			wantToDelete: []string{},
		},
		{
			name: "unchanged file — same timestamp",
			filesFound: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": old,
			},
			snapshots: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": old,
			},
			wantToIndex:  []string{},
			wantToDelete: []string{},
		},
		{
			name: "mixed — one new, one deleted, one unchanged",
			filesFound: map[string]time.Time{
				"/manga/Berserk/Chapter_057/014.jpg": now, // new
				"/manga/Berserk/Chapter_057/015.jpg": old, // unchanged
			},
			snapshots: map[string]time.Time{
				"/manga/Berserk/Chapter_057/015.jpg": old, // unchanged
				"/manga/Berserk/Chapter_057/016.jpg": old, // deleted
			},
			wantToIndex:  []string{"/manga/Berserk/Chapter_057/014.jpg"},
			wantToDelete: []string{"/manga/Berserk/Chapter_057/016.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				filesFound: tt.filesFound,
				mainFolder: "/manga",
				stopCh:     make(chan struct{}),
			}

			db := &mockDB{snapshots: tt.snapshots}
			toIndex, toDelete, err := w.compareWithoutScan(context.Background(), db)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			sort.Strings(toIndex)
			sort.Strings(toDelete)
			sort.Strings(tt.wantToIndex)
			sort.Strings(tt.wantToDelete)

			if len(toIndex) != len(tt.wantToIndex) {
				t.Errorf("toIndex: got %v, want %v", toIndex, tt.wantToIndex)
			} else {
				for i := range toIndex {
					if toIndex[i] != tt.wantToIndex[i] {
						t.Errorf("toIndex[%d]: got %q, want %q", i, toIndex[i], tt.wantToIndex[i])
					}
				}
			}

			if len(toDelete) != len(tt.wantToDelete) {
				t.Errorf("toDelete: got %v, want %v", toDelete, tt.wantToDelete)
			} else {
				for i := range toDelete {
					if toDelete[i] != tt.wantToDelete[i] {
						t.Errorf("toDelete[%d]: got %q, want %q", i, toDelete[i], tt.wantToDelete[i])
					}
				}
			}
		})
	}
}