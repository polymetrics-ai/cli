package connsdk

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestMultipartFilenamePolicy217(t *testing.T) {
	for _, tc := range []struct {
		name, policy, want string
		invalid            bool
	}{
		{"payload.bin", "url_percent_utf8", "payload.bin", false},
		{"résumé.bin", "url_percent_utf8", "r%C3%A9sum%C3%A9.bin", false},
		{"two words.bin", "url_percent_utf8", "two%20words.bin", false},
		{"a+b.bin", "url_percent_utf8", "a%2Bb.bin", false},
		{"100%.bin", "url_percent_utf8", "100%25.bin", false},
		{"r%C3%A9.bin", "url_percent_utf8", "r%25C3%25A9.bin", false},
		{"résumé.bin", "", "résumé.bin", false},
		{"payload.bin", "unknown", "", true},
		{"../payload.bin", "url_percent_utf8", "", true},
		{"bad\nname.bin", "url_percent_utf8", "", true},
	} {
		t.Run(tc.name+tc.policy, func(t *testing.T) {
			source := filepath.Join(t.TempDir(), "source.bin")
			payload := []byte("independent-file-217")
			if err := os.WriteFile(source, payload, 0600); err != nil {
				t.Fatal(err)
			}
			var mu sync.Mutex
			var bodies [][]byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(io.LimitReader(r.Body, 8193))
				if err != nil {
					t.Error(err)
				}
				mu.Lock()
				bodies = append(bodies, raw)
				mu.Unlock()
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()
			requester := Requester{BaseURL: server.URL, Client: server.Client()}
			form := MultipartForm{MaxBytes: 8192, Files: []MultipartFile{{FieldName: "file", Path: source, FileName: tc.name, FilenameEncoding: tc.policy, MaxBytes: 1024, ContentType: "application/octet-stream"}}}
			_, err := requester.DoMultipartLimited(t.Context(), "POST", "/upload", nil, form, 1024)
			mu.Lock()
			defer mu.Unlock()
			if tc.invalid {
				if err == nil || len(bodies) != 0 {
					t.Fatalf("invalid filename policy sent: error=%v sends=%d", err, len(bodies))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(bodies) != 1 || !bytes.Contains(bodies[0], payload) {
				t.Fatalf("file body changed or wrong sends=%d", len(bodies))
			}
			header := `Content-Disposition: form-data; name="file"; filename="` + tc.want + `"` + "\r\n"
			if !bytes.Contains(bodies[0], []byte(header)) || bytes.Contains(bodies[0], []byte("filename*=")) {
				t.Fatalf("raw header missing literal %q in %q", header, bodies[0])
			}
		})
	}
}

func TestMultipartMetadataBudget217(t *testing.T) {
	source := filepath.Join(t.TempDir(), "payload.bin")
	payload := []byte("file-217")
	if err := os.WriteFile(source, payload, 0600); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var bodies [][]byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(io.LimitReader(r.Body, 8193))
		if err != nil {
			t.Error(err)
		}
		mu.Lock()
		bodies = append(bodies, raw)
		mu.Unlock()
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	requester := Requester{BaseURL: server.URL, Client: server.Client()}
	form := MultipartForm{Fields: map[string]string{"parent": "parent-217"}, MaxBytes: 8192, Files: []MultipartFile{{FieldName: "file", Path: source, MaxBytes: 1024, ContentType: "application/octet-stream"}}}
	if _, err := requester.DoMultipartLimited(t.Context(), "POST", "/upload", nil, form, 1024); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	if len(bodies) != 1 || bytes.Count(bodies[0], payload) != 1 {
		t.Fatal("independent metadata partition fixture invalid")
	}
	metadata := int64(len(bodies[0]) - len(payload))
	mu.Unlock()
	form.MaxMetadataBytes = &metadata
	if _, err := requester.DoMultipartLimited(t.Context(), "POST", "/upload", nil, form, 1024); err != nil {
		t.Fatalf("exact measured metadata refused: %v", err)
	}
	form.Fields["parent"] += "x"
	if _, err := requester.DoMultipartLimited(t.Context(), "POST", "/upload", nil, form, 1024); err == nil || !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("metadata one-over accepted: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 {
		t.Fatalf("over-budget provider sends=%d want2", len(bodies))
	}
}

func TestMultipartExactFileEnvelope217(t *testing.T) {
	const fileLimit int64 = 104857600
	source := filepath.Join(t.TempDir(), "résumé +%.bin")
	f, err := os.Create(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(fileLimit); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var sends int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		sends++
		mu.Unlock()
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			return
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		if part.FormName() != "file" || part.FileName() != "r%C3%A9sum%C3%A9%20%2B%25.bin" {
			t.Errorf("snapshot wire identity %q %q", part.FormName(), part.FileName())
		}
		digest := sha256.New()
		n, err := io.Copy(digest, io.LimitReader(part, fileLimit+1))
		if err != nil {
			t.Error(err)
		}
		if n != fileLimit || hex.EncodeToString(digest.Sum(nil)) != "20492a4d0d84f8beb1767f6616229f85d44c2827b64bdbfb260ee12fa1109e0e" {
			t.Errorf("file count/digest mismatch %d", n)
		}
		if _, err = reader.NextPart(); err != io.EOF {
			t.Errorf("unexpected next part: %v", err)
		}
		_, _ = w.Write([]byte(`{"stored_bytes":104857600}`))
	}))
	defer server.Close()
	budget := int64(65536)
	form := MultipartForm{MaxBytes: 104923136, MaxMetadataBytes: &budget, Files: []MultipartFile{{FieldName: "file", Path: source, MaxBytes: fileLimit, FilenameEncoding: "url_percent_utf8", ContentType: "application/octet-stream", ExpectedSHA256: "20492a4d0d84f8beb1767f6616229f85d44c2827b64bdbfb260ee12fa1109e0e"}}}
	requester := Requester{BaseURL: server.URL, Client: server.Client()}
	result, err := requester.DoMultipartLimited(t.Context(), "POST", "/upload", nil, form, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if string(result.Body) != `{"stored_bytes":104857600}` {
		t.Fatalf("returned outcome %q", result.Body)
	}
	if err = os.Truncate(source, fileLimit+1); err != nil {
		t.Fatal(err)
	}
	if _, err = requester.DoMultipartLimited(t.Context(), "POST", "/upload", nil, form, 1024); err == nil {
		t.Fatal("file cap+1 accepted")
	}
	mu.Lock()
	defer mu.Unlock()
	if sends != 1 {
		t.Fatalf("sends %d", sends)
	}
}
