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

type httpHandler struct {
	Dst string
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

	hostPort := fmt.Sprintf("127.0.0.1:%d", port)
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
		if len(r.URL.String()) > 0 {
			if !strings.HasSuffix(redir, "/") && scriptMatch(redir) == false {
				redir = fmt.Sprintf("%s/", redir)
			}
			redir = fmt.Sprintf("%s?url=%s", redir, r.URL.String())
		}
	}

	http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
}

func scriptMatch(path string) bool {
	end := ".(php|pl|py|cgi|do|html|phtml)$"
	match, _ := regexp.MatchString(end, path)
	return match
}

func uaBlackList(ua string) bool {
	list := []string{
		"^BTWebClient",
		"^uTorrent",
	}

	for _, name := range list {
		if match, _ := regexp.MatchString(name, ua); match {
			return true
		}
	}

	return false
}
