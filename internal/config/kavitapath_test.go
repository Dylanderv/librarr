package config

import "testing"

func TestResolveKavitaPath(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "shared mount under manga root",
			cfg:  Config{KavitaMangaLibraryPath: "/data/manga"},
			in:   "/data/manga/One Piece/vol1.cbz",
			want: "/data/manga/One Piece/vol1.cbz",
		},
		{
			name: "remap prefix",
			cfg:  Config{KavitaPathMap: "/manga:/data/manga"},
			in:   "/manga/Bleach/v1.cbz",
			want: "/data/manga/Bleach/v1.cbz",
		},
		{
			name:    "traversal escape rejected",
			cfg:     Config{KavitaMangaLibraryPath: "/data/manga"},
			in:      "/data/manga/../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "outside roots rejected",
			cfg:     Config{KavitaMangaLibraryPath: "/data/manga"},
			in:      "/other/place/book.cbz",
			wantErr: true,
		},
		{
			name:    "no roots configured",
			cfg:     Config{},
			in:      "/data/manga/x.cbz",
			wantErr: true,
		},
		{
			name: "sibling prefix not treated as under root",
			cfg:  Config{KavitaMangaLibraryPath: "/data/manga"},
			in:   "/data/manga-extra/x.cbz",
			// /data/manga-extra should NOT be considered under /data/manga
			wantErr: true,
		},
		{
			name: "remap picks correct pair among many",
			cfg:  Config{KavitaPathMap: "/manga:/data/manga,/books:/data/books"},
			in:   "/books/Author/Title.epub",
			want: "/data/books/Author/Title.epub",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cfg.ResolveKavitaPath(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
