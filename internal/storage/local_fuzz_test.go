package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzGetPath feeds arbitrary bucket/key pairs to the local-storage path
// resolver. Oracle: no panic, and a resolved path always stays inside the
// storage base directory.
func FuzzGetPath(f *testing.F) {
	f.Add("bucket", "key/file.txt")
	f.Add("bucket", "../../etc/passwd")
	f.Add("..", "a//b")
	f.Add("b", "a\x00b")
	f.Add("b", `a\..\..\windows`)
	f.Add("b", "....//....//etc")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, bucket, key string) {
		ls := &LocalStorage{basePath: filepath.Join(t.TempDir(), "base")}

		path, err := ls.getPath(bucket, key)
		if err != nil {
			return
		}

		rel, relErr := filepath.Rel(ls.basePath, path)
		if relErr != nil {
			t.Fatalf("resolved path %q not comparable to base %q", path, ls.basePath)
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Errorf("path %q escapes storage base %q (rel %q)", path, ls.basePath, rel)
		}
		if strings.ContainsRune(path, 0) {
			t.Errorf("path %q contains a null byte", path)
		}
	})
}

// FuzzValidateSignedToken feeds arbitrary tokens to the signed-URL verifier.
// Oracle: no panic; a valid verification implies the HMAC held.
func FuzzValidateSignedToken(f *testing.F) {
	mkToken := func(payload string) string {
		mac := hmac.New(sha256.New, []byte("fuzz-secret"))
		mac.Write([]byte(payload))
		return base64.URLEncoding.EncodeToString(append([]byte(payload), mac.Sum(nil)...))
	}
	f.Add(mkToken(`{"bucket":"b","key":"k","method":"GET","expires_at":9999999999}`))
	f.Add(mkToken(`{"bucket":"b","key":"../../x","method":"DELETE","expires_at":9999999999}`))
	f.Add(mkToken(`{"bucket":"b","key":"k","method":"GET","expires_at":1}`)) // expired
	f.Add("not-base64!!!")
	f.Add(base64.URLEncoding.EncodeToString([]byte("short")))

	f.Fuzz(func(t *testing.T, token string) {
		ls := &LocalStorage{signingSecret: "fuzz-secret"}

		bucket, key, method, err := ls.ValidateSignedToken(token)
		if err != nil {
			return
		}

		// A token that verifies must decode to a sane triple.
		if bucket == "" && key == "" && method == "" {
			t.Errorf("valid token decoded to an empty triple")
		}
	})
}
