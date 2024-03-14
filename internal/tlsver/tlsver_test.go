package tlsver_test

import (
	"testing"

	"github.com/jreisinger/tools/internal/tlsver"
)

func TestGetGetsTLSVersionOfCloudflareServer(t *testing.T) {
	t.Parallel()
	g := tlsver.NewGetter("1.1.1.1", "443")
	if g.Get(); g.Err != nil {
		t.Error(g.Err)
	}
	want := "1.3"
	got := g.TLSversion.String()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}
