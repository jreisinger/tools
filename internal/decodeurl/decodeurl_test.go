package decodeurl_test

import (
	"testing"

	"github.com/jreisinger/tools/internal/decodeurl"

	"github.com/google/go-cmp/cmp"
)

func TestDecode_DecodesURL(t *testing.T) {
	t.Parallel()
	URL := "https://example.com/some%21/path?key1=value1"
	got, err := decodeurl.Decode(URL)
	if err != nil {
		t.Fatal(err)
	}
	want := decodeurl.DecodedURL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/some!/path",
		QueryKeyValuePairs: map[string][]string{
			"key1": {"value1"},
		},
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
