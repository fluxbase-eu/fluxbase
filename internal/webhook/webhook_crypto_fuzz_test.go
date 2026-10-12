package webhook

import (
	"testing"
	"time"
)

// FuzzParseWebhookSignature feeds arbitrary X-Fluxbase-Signature headers to
// the parser and verifier. Oracle: no panic; a nil error always yields a
// usable signature (non-zero timestamp, at least one signature value).
func FuzzParseWebhookSignature(f *testing.F) {
	f.Add("t=1614556800,v1=5257a869e7ecebeda32affa62cdca3fa51cad7e77a0e56ff536d0ce8e108d8bd", "payload", int64(300))
	f.Add("t=abc,v1=x", "payload", int64(300))
	f.Add("garbage", "payload", int64(300))
	f.Add("t=1,v1=a,v1=b,v1=c", "payload", int64(0))
	f.Add("t=99999999999999999999,v1=x", "payload", int64(-1))

	f.Fuzz(func(t *testing.T, header, payload string, toleranceSecs int64) {
		sig, err := ParseWebhookSignature(header)
		if err != nil {
			if sig != nil {
				t.Errorf("ParseWebhookSignature returned error %v with non-nil signature", err)
			}
			return
		}
		if sig == nil || sig.Timestamp == 0 || len(sig.Signatures) == 0 {
			t.Errorf("success must yield a non-zero timestamp and at least one signature (got %+v)", sig)
		}

		// The verifier must also be panic-free on arbitrary inputs.
		_ = VerifyWebhookSignature([]byte(payload), header, "fuzz-secret", time.Duration(toleranceSecs)*time.Second)
	})
}
