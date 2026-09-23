package spider

import (
	"testing"
)

func TestCrawlA(t *testing.T) {
	x, err := CrawlA("http://192.168.0.128/programs/", 1, 0)
	t.Log(x)
	t.Log(err)
}
