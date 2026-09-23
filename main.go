package main

import (
	"crypto/sha256"
	"crypto/sha3"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/containedreality/nekoget/constants"
	"github.com/containedreality/nekoget/progresstracker"
	"github.com/containedreality/nekoget/server"
	"github.com/containedreality/nekoget/spider"
)

type DownloadJob struct {
	URL      string
	Filename string
}

func DownloadHttp(url string, outname string, progressbar bool) {
	buf := make([]byte, 1024*1024)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", constants.USER_AGENT)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("%s got a %d.\n", url, resp.StatusCode)
		return
	}

	var out *os.File

	switch outname {
	case "":
		outname = filepath.Base(url)
		out, err = os.Create(outname)
	case "-":
		out = os.Stdout
	default:
		out, err = os.Create(outname)

	}

	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()
	defer resp.Body.Close()

	// sha2-256 is only really here cause most things use it.
	sha2_256 := sha256.New()
	sha3_512 := sha3.New512()

	var writers io.Writer

	if progressbar {
		x := &progresstracker.ProgressTracker{
			Underlying: out,
			Total:      resp.ContentLength,
		}

		writers = io.MultiWriter(x, sha2_256, sha3_512)
	} else {
		writers = io.MultiWriter(out, sha2_256, sha3_512)
	}

	_, err = io.CopyBuffer(writers, resp.Body, buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stderr, "\r\033[K")
	if outname == "-" {
		log.Printf("Downloaded: %s (stdout)\n", url)
	} else {
		log.Printf("Downloaded: %s\n", outname)
	}
	log.Printf("SHA2_256: %x\n", sha2_256.Sum(nil))
	log.Printf("SHA3_512: %x\n", sha3_512.Sum(nil))
}

func downloadListWorker(jobs <-chan DownloadJob, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		DownloadHttp(job.URL, job.Filename, false)
		time.Sleep(time.Millisecond * 250)
	}
}

func DownloadHttpList(file string, dirs bool, threads int) {
	var wg sync.WaitGroup

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	str := string(b)
	str = strings.ReplaceAll(str, "\r", "")

	urls := strings.Split(str, "\n")
	jobs := make(chan DownloadJob, len(urls))

	for range threads {
		wg.Add(1)
		go downloadListWorker(jobs, &wg)
	}

	for _, u := range urls {
		if u == "" {
			continue
		}

		parsed, err := url.Parse(u)
		if err != nil {
			log.Printf("url.Parse: %s\n", err)
		}

		out := parsed.Path[1:]
		dir := filepath.Dir(out)

		fname := ""

		if dirs {
			err = os.MkdirAll(dir, 0755)

			if err != nil {
				log.Printf("os.MkdirAll: %s\n", err)
				continue
			}

			fname = out
		}

		jobs <- DownloadJob{URL: u, Filename: fname}
	}

	close(jobs)
	wg.Wait()
}

func Spider(url string, out string, maxdepth int) {
	nameset := false

	urls, err := spider.CrawlA(url, maxdepth, 0)
	if err != nil {
		log.Fatal(err)
	}

	f := os.Stdout

	if out != "" && out != "-" {
		f, err = os.Create(out)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		nameset = true
	}

	for _, s := range urls {
		fmt.Fprintf(f, "%s\n", s)
	}

	if nameset {
		log.Printf("%s saved to %s\n", url, out)
	}
}

func main() {
	banner := `

   m                         #                             m
 m"#"m         m mm    mmm   #   m   mmm    mmmm   mmm   mm#mm
 #m#           #"  #  #"  #  # m"   #" "#  #" "#  #"  #    #
   #"#         #   #  #""""  #"#    #   #  #   #  #""""    #
 "m#m"         #   #  "#mm"  #  "m  "#m#"  "#m"#  "#mm"    "mm
   #                                        m  #
       """"""                                ""
	`

	fmt.Fprintln(os.Stderr, banner)

	var spiderFlag bool
	var dirs bool
	var maxdepth int
	var threads int
	var list string
	var out string
	var path string
	var host string

	flag.StringVar(&list, "list", "", "a list of URLs to download from")
	flag.StringVar(&out, "out", "", "the file to output to, use - for stdout.")
	flag.StringVar(&path, "serve-path", "", "the path to serve, must be set to enable server mode.")
	flag.StringVar(&host, "serve-host", ":8000", "host to serve files on")
	flag.IntVar(&threads, "threads", 4, "threads for downloading from a list")
	flag.BoolVar(&spiderFlag, "spider", false, "creates a list of urls that can then be downloaded. use with caution as it's likely to cause issues on some services, it also doesn't work fully.")
	flag.BoolVar(&dirs, "dirs", false, "create directories when downloading from a list")
	flag.IntVar(&maxdepth, "maxdepth", 0, "maximum depth for the spider to crawl, use 0 for no limit on depth.")

	flag.Parse()

	urls := flag.Args()

	if len(urls) == 0 && list == "" && path == "" {
		flag.Usage()
		log.Fatal("you gotta provide a url/list to download from or just a URL if you want to spider.")
	} else if path != "" {
		if err := server.ServeFiles(host, path); err != nil {
			log.Fatalln(err)
		}
	}

	if list != "" {
		DownloadHttpList(list, dirs, threads)
	} else if spiderFlag {
		for _, url := range urls {
			Spider(url, out, maxdepth)
		}
	} else {
		for _, url := range urls {
			DownloadHttp(url, out, true)
		}
	}
}
