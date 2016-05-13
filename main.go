package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	uaList   []*regexp.Regexp
	isScript *regexp.Regexp
)

type httpHandler struct {
	Dst string
}

func init() {
	var err error

	list := []string{
		"^(BTWebClient|uTorrent|Microsoft-CryptoAPI|Windows-Update-Agent|Microsoft BITS|Google Update|GoogleEarth|MRA|MediaGet|Syncer|Akamai)",
		"^(Skype|Avast|avast|Apache-HttpClient|SCSDK)",
		"^(MSDW)$",
		"(DrWebUpdate|MailRuSputnik|Microsoft NCSI|Windows-Update-Agent|BenchHttp)",
		"^$", // empty
	}
	for _, line := range list {
		r, err := regexp.Compile(line)
		if err != nil {
			log.Panic(err)
		}
		uaList = append(uaList, r)
	}

	isScript, err = regexp.Compile(".(php|pl|py|cgi|fcgi|do|htm|html|phtml|xhtml)$")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	var wg sync.WaitGroup
	for port, dst := range discover() {
		wg.Add(1)
		go runWeb(port, dst, &wg)
	}
	wg.Wait()
}

func discover() map[int]string {
	cap := map[int]string{}

	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "CAPTIVE") {
			keys := strings.SplitN(kv, "=", 2)
			mkey := strings.SplitN(keys[0], "_", 2)

			if len(mkey[1]) >= 2 && len(mkey[1]) <= 5 {
				if len(mkey) == 2 {
					port, err := strconv.Atoi(mkey[1])
					if err != nil {
						log.Panic(err)
					}
					cap[port] = keys[1]
				}
			}

		}
	}

	return cap
}

func runWeb(port int, dst string, wg *sync.WaitGroup) {
	defer wg.Done()

	hostPort := fmt.Sprintf("0.0.0.0:%d", port)
	log.Print("ListenAndServe: ", hostPort, " -> ", dst)

	srv := &http.Server{
		Addr:           hostPort,
		Handler:        httpHandler{Dst: dst},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1mb
	}
	srv.SetKeepAlivesEnabled(false)

	log.Fatal("ListenAndServe: ", srv.ListenAndServe())
}

func (this httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if uaBlackList(r.UserAgent()) {
		w.Header().Add("Retry-After", fmt.Sprintf("%d", 5*60)) // 5min
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	redir := this.Dst

	if len(os.Getenv("CAPTIVE_NOORIGIN")) == 0 {
		if len(r.URL.String()) > 7 {
			if !strings.HasSuffix(redir, "/") && !isScript.MatchString(redir) {
				redir = fmt.Sprintf("%s/", redir)
			}
			scheme := r.URL.Scheme
			if scheme == "" {
				scheme = "http"
			}
			redir = fmt.Sprintf("%s?url=%s://%s%s", redir, scheme, r.Host, r.URL.String())
		}
	}

	http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
}

func uaBlackList(ua string) bool {
	for _, name := range uaList {
		if match := name.MatchString(ua); match {
			return true
		}
	}
	return false
}
