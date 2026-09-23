package spider

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/containedreality/nekoget/constants"
)

// crawl all <a> tags
// doesn't respect robots.txt that means use with caution.
// works for nginx anything else idk i haven't tested nor do i care much to
func CrawlA(aurl string, maxdepth int, depth int) ([]string, error) {
	if maxdepth != 0 {
		depth++

		if depth > maxdepth {
			return nil, nil
		}
	}

	var urls []string

	req, err := http.NewRequest(http.MethodGet, aurl, nil)
	req.Header.Set("User-Agent", constants.USER_AGENT)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		data, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		str := strings.ReplaceAll(string(data), "\r", "")

		for v := range strings.SplitSeq(str, "\n") {
			if strings.Contains(v, "<a href=\"") {
				x := strings.Split(v, "<a href=\"")[1]
				x = strings.Split(x, "\">")[0]

				if x == "../" || x[0] == '#' {
					continue
				}

				f := "%s/%s"

				if aurl[len(aurl)-1] == '/' {
					f = "%s%s"
				}

				u := fmt.Sprintf(f, aurl, x)

				if x[len(x)-1] == '/' {
					ourls, err := CrawlA(u, maxdepth, depth)
					if err != nil {
						continue
					}

					urls = append(urls, ourls...)
				} else {
					urls = append(urls, u)
				}
			}
		}
	}

	return urls, nil
}
