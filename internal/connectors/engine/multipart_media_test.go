package engine

import (
	"strings"
	"testing"
)

func TestValidateMultipartMediaTypes(t *testing.T) {
	tests := []struct {
		name    string
		part    MultipartPartSpec
		wantErr string
	}{
		{
			name: "absent is unconstrained",
			part: MultipartPartSpec{Name: "file", Type: "file", Field: "path"},
		},
		{
			name: "declared list is accepted",
			part: MultipartPartSpec{Name: "file", Type: "file", Field: "path", ContentType: "image/png", AllowedMediaTypes: []string{"image/png", "image/jpeg"}},
		},
		{
			name:    "present but empty is refused",
			part:    MultipartPartSpec{Name: "file", Type: "file", Field: "path", AllowedMediaTypes: []string{}},
			wantErr: "must not be empty",
		},
		{
			name:    "unparseable entry is refused",
			part:    MultipartPartSpec{Name: "file", Type: "file", Field: "path", AllowedMediaTypes: []string{"not a media type"}},
			wantErr: "not a valid media type",
		},
		{
			name:    "content_type outside its own allowlist is refused",
			part:    MultipartPartSpec{Name: "file", Type: "file", Field: "path", ContentType: "application/zip", AllowedMediaTypes: []string{"image/png"}},
			wantErr: "not among its own allowed_media_types",
		},
		{
			name:    "media bound on a non-file part is refused",
			part:    MultipartPartSpec{Name: "meta", Type: "field", Field: "meta", AllowedMediaTypes: []string{"image/png"}},
			wantErr: "only meaningful on a file part",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMultipartMediaTypes(tt.part)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateMultipartMediaTypes error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateMultipartMediaTypes error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// TestMultipartRootRelativePath pins how record-supplied paths are converted for
// os.Root. Absolute paths inside the project directory stay accepted, as they
// were before confinement, and anything escaping is refused.
func TestMultipartRootRelativePath(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "relative stays relative", raw: "media/clip.mp4", want: "media/clip.mp4"},
		{name: "absolute inside the root is made relative", raw: dir + "/media/clip.mp4", want: "media/clip.mp4"},
		{name: "relative traversal is refused", raw: "../outside.txt", wantErr: true},
		{name: "absolute outside the root is refused", raw: "/etc/passwd", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := multipartRootRelativePath(dir, tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("multipartRootRelativePath(%q) = %q, want an error", tt.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("multipartRootRelativePath(%q) error = %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("multipartRootRelativePath(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
