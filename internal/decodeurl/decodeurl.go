package decodeurl

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
)

type DecodedURL struct {
	Scheme string
	Host   string
	Path   string
	QueryKeyValuePairs
}

type QueryKeyValuePairs map[string][]string

func (du DecodedURL) String() string {
	return fmt.Sprintf(`
Scheme	%s 
Host	%s
Path	%s
Query
%s`,
		du.Scheme, du.Host, du.Path, parseQuery(du.QueryKeyValuePairs))
}

func parseQuery(q QueryKeyValuePairs) string {
	var output []string

	// Sort map by keys.
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		output = append(output, fmt.Sprintf("\t%s = %s", k, q[k]))
	}
	return strings.Join(output, "\n")
}

func Decode(URL string) (DecodedURL, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return DecodedURL{}, err
	}
	du := DecodedURL{
		Scheme:             u.Scheme,
		Host:               u.Host,
		Path:               u.Path,
		QueryKeyValuePairs: QueryKeyValuePairs(u.Query()),
	}
	return du, nil
}

const usage = `Usage: decodeurl <url>

Example: decodeurl 'https://example.com/some%21/path?key1=value1,value2&key2=abc%2B%2B'
`

func Main() {
	if len(os.Args) != 2 {
		fmt.Print(usage)
		os.Exit(1)
	}
	du, err := Decode(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(du)
}
