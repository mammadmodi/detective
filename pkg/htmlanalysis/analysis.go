package htmlanalysis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/html"
)

// HeadingsCount is a type which holds headings count by their level.
type HeadingsCount struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

// LinksCount is a struct which holds links count based on their region.
type LinksCount struct {
	Internal int `json:"internal"`
	External int `json:"external"`
}

// Result is a report of analysis on a html document.
// HTMLVersion is version of HTML which is mentioned in doctype tag.
// PageTitle is title of page in the title tag.
// HeadingsCount is count of headings by their level.
// InaccessibleLinksCount is count of links that doesn't return a 2xx status code
// on a GET request.
// HasLoginForm shows that whether the html doc contains a login form or not.
type Result struct {
	HTMLVersion            string         `json:"html_version"`
	PageTitle              string         `json:"page_title"`
	HeadingsCount          *HeadingsCount `json:"headings_count"`
	LinksCount             *LinksCount    `json:"links_count"`
	InaccessibleLinksCount int            `json:"inaccessible_links_count"`
	HasLoginForm           bool           `json:"has_login_form"`
}

var globalLogger = zap.NewNop()

// SetGlobalLogger sets a logger for this package.
func SetGlobalLogger(logger *zap.Logger) {
	globalLogger = logger
}

var globalHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		// We should reduce the IdleConnTimeout because the requests that are being performed
		// by this HTTPClient target different hosts and there is no meaning to have idle connection
		// for a long time.
		IdleConnTimeout: 5 * time.Second,
	},
}

// SetGlobalHTTPClient sets an HTTPClient for this package which is used for analysis purposes.
func SetGlobalHTTPClient(client *http.Client) {
	globalHTTPClient = client
}

// Analyze is a global wrapper function on htmlAnalyzer.Analyze.
func Analyze(ctx context.Context, hostURL *url.URL, htmlDocument string) (*Result, error) {
	h := newHTMLAnalyzer(htmlDocument, hostURL)
	return h.analyze(ctx)
}

// newHTMLAnalyzer creates a new htmlAnalyzer object.
func newHTMLAnalyzer(htmlDoc string, hostURL *url.URL) *htmlAnalyzer {
	return &htmlAnalyzer{
		htmlDoc:       htmlDoc,
		hostURL:       hostURL,
		internalLinks: []*url.URL{},
		externalLinks: []*url.URL{},
	}
}

// htmlAnalyzer is a struct which holds the states of result during the analysis.
type htmlAnalyzer struct {
	htmlDoc       string
	hostURL       *url.URL
	result        *Result
	internalLinks []*url.URL
	externalLinks []*url.URL
}

// analyze starts analysis on htmlAnalyzer.htmlDoc field.
func (h *htmlAnalyzer) analyze(ctx context.Context) (*Result, error) {
	if h.result != nil {
		return h.result, nil
	}

	if err := h.validate(); err != nil {
		return nil, errors.New("html document is not valid")
	}

	h.result = &Result{}
	h.
		parseAndSetHTMLVersion().
		parseAndSetPageTitle().
		parseAndSetHeadingsCount().
		parseAndSetLinksCount().
		setInaccessibleLinksCount(ctx).
		parseAndSetHasLoginForm()

	return h.result, nil
}

// validate loops on all of html tags and checks any problems in that.
func (h *htmlAnalyzer) validate() error {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDoc))
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error while tokenizing HTML: %v", err)
		}
	}
}

// parseAndSetHTMLVersion parses html document and sets the version in result field.
func (h *htmlAnalyzer) parseAndSetHTMLVersion() *htmlAnalyzer {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDoc))
	version := "Unknown HTML Version"

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		if tt == html.DoctypeToken {
			type docType struct {
				version, matcher string
			}
			// This is a list of available html versions which is been recommended by `www.w3.org`.
			// https://www.w3.org/QA/2002/04/valid-dtd-list.html
			docTypes := [8]docType{
				{version: "HTML 4.01 Strict", matcher: `"-//W3C//DTD HTML 4.01//EN"`},
				{version: "HTML 4.01 Transitional", matcher: `"-//W3C//DTD HTML 4.01 TRANSITIONAL//EN"`},
				{version: "HTML 4.01 Frameset", matcher: `"-//W3C//DTD HTML 4.01 FRAMESET//EN"`},
				{version: "XHTML 1.0 Strict", matcher: `"-//W3C//DTD XHTML 1.0 STRICT//EN`},
				{version: "XHTML 1.0 Transitional", matcher: `"-//W3C//DTD XHTML 1.0 TRANSITIONAL//EN"`},
				{version: "XHTML 1.0 Frameset", matcher: `"-//W3C//DTD XHTML 1.0 FRAMESET//EN"`},
				{version: "XHTML 1.1", matcher: `"-//W3C//DTD XHTML 1.1//EN"`},
				{version: "HTML 5", matcher: `HTML`},
			}
			v := tokenizer.Token().Data
			for _, d := range docTypes {
				ok := strings.Contains(strings.ToUpper(v), d.matcher)
				if ok {
					version = d.version
					break
				}
			}
		}
	}
	h.result.HTMLVersion = version
	return h
}

// parseAndSetPageTitle parses html document and sets the page title in result field.
func (h *htmlAnalyzer) parseAndSetPageTitle() *htmlAnalyzer {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDoc))
	pageTitle := "Empty Page Title"
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		td := tokenizer.Token().Data
		if tt == html.StartTagToken && td == "title" {
			tt = tokenizer.Next()
			if tt == html.TextToken {
				pageTitle = tokenizer.Token().Data
			}
			break
		}
	}
	h.result.PageTitle = pageTitle
	return h
}

// parseAndSetHeadingsCount parses html doc and sets headings based on their levels to result field.
func (h *htmlAnalyzer) parseAndSetHeadingsCount() *htmlAnalyzer {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDoc))
	headings := &HeadingsCount{}
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		td := tokenizer.Token().Data
		if tt == html.StartTagToken && string(td[0]) == "h" {
			switch td {
			case "h1":
				headings.H1++
			case "h2":
				headings.H2++
			case "h3":
				headings.H3++
			case "h4":
				headings.H4++
			case "h5":
				headings.H5++
			case "h6":
				headings.H6++
			}
		}
	}
	h.result.HeadingsCount = headings
	return h
}

// parseAndSetLinksCount first sets all the available links in the state of htmlAnalyzer
// and then sets the LinksCount field to result.
func (h *htmlAnalyzer) parseAndSetLinksCount() *htmlAnalyzer {
	h.parseAndSetLinks()
	h.result.LinksCount = &LinksCount{
		Internal: len(h.internalLinks),
		External: len(h.externalLinks),
	}
	return h
}

// parseAndSetLinks parses the html document and stores all the internal and external links
// to htmlAnalyzer.internalLinks and htmlAnalyzer.externalLinks.
// we store this links because we need them for finding inaccessible links count.
func (h *htmlAnalyzer) parseAndSetLinks() {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDoc))
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			return
		}

		t := tokenizer.Token()
		if tt == html.StartTagToken && t.Data == "a" {
			for _, attr := range t.Attr {
				if attr.Key == "href" {
					// Empty href.
					if len(attr.Val) == 0 {
						globalLogger.With(zap.String("href", attr.Val)).Debug("url ignored because it was empty")
						continue
					}
					// Pointer links.
					if string(attr.Val[0]) == "#" {
						globalLogger.With(zap.String("href", attr.Val)).Debug("url ignored because it was a pointer")
						continue
					}

					u, err := url.Parse(strings.TrimSpace(attr.Val))
					if err != nil {
						globalLogger.With(zap.String("href", attr.Val)).Debug("url ignored, could not Parse url")
						continue
					}

					if u.Scheme != "" && !strings.Contains(u.Scheme, "http") {
						globalLogger.With(zap.String("href", attr.Val), zap.String("scheme", u.Scheme)).
							Debug("url ignored, bad scheme")
						continue
					}

					if h.isInternalLink(u) {
						u = u.ResolveReference(h.hostURL)
						h.internalLinks = append(h.internalLinks, u)
						globalLogger.With(zap.String("url", u.String())).Info("marked as internal")
					} else {
						h.externalLinks = append(h.externalLinks, u)
						globalLogger.With(zap.String("url", u.String())).Info("marked as external")
					}
				}
			}
		}
	}
}

func (h *htmlAnalyzer) isInternalLink(url *url.URL) bool {
	return url.Host == "" || strings.Contains(strings.ToLower(url.Host), h.hostURL.Host)
}

// setInaccessibleLinksCount loops on all of links and counts the links that doesn't return
// an acceptable 2xx status code.
func (h *htmlAnalyzer) setInaccessibleLinksCount(ctx context.Context) *htmlAnalyzer {
	var m sync.Mutex
	var inaccessibleLinksCount int
	inc := func() {
		m.Lock()
		inaccessibleLinksCount++
		m.Unlock()
	}

	totalLinks := append(h.externalLinks, h.internalLinks...)
	wg := sync.WaitGroup{}
	wg.Add(len(totalLinks))
	for _, u := range totalLinks {
		u := u
		go func() {
			defer wg.Done()
			if !h.isAccessibleURL(ctx, u) {
				globalLogger.With(zap.String("url", u.String())).Debug("url is not accessible")
				inc()
				return
			}
			globalLogger.With(zap.String("url", u.String())).Debug("url is accessible")
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		globalLogger.Debug("all go routines finished successfully")
		done <- struct{}{}
	}()

	select {
	case <-done:
		// here we have finished process of inaccessible links before ending of context.
	case <-ctx.Done():
		globalLogger.Error("process stopped due to context got done")
	}
	h.result.InaccessibleLinksCount = inaccessibleLinksCount
	return h
}

// isAccessibleURL checks the accessibility of a link.
func (h *htmlAnalyzer) isAccessibleURL(ctx context.Context, u *url.URL) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		globalLogger.Error("could not create request")
		return false
	}

	resp, err := globalHTTPClient.Do(req)
	if err != nil {
		globalLogger.Error("could not perform request")
		return false
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	}

	globalLogger.Error("response code is not 200")
	return false
}

// parseAndSetHasLoginForm parses the document and sets a flag in result field.
func (h *htmlAnalyzer) parseAndSetHasLoginForm() *htmlAnalyzer {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDoc))
	hasLoginForm := false
	// If the html has a form which has one of the following keywords in it's identity attributes
	// we can be sure that page has login form.
	loginFormKeywords := []string{"login", "signin", "sign_in"}
	// If the html has password inputs only once we can say that it's a login form.
	// If number of password inputs is more than 1 so it's a sign up page or a reset password.
	var numOfPasswordInputs uint8
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			hasLoginForm = numOfPasswordInputs == 1
			break
		}
		t := tokenizer.Token()
		if tt == html.StartTagToken {
			if t.Data == "form" {
				for _, attr := range t.Attr {
					if attr.Key == "name" || attr.Key == "id" || attr.Key == "action" {
						for _, k := range loginFormKeywords {
							if strings.Contains(strings.ToLower(attr.Val), k) {
								hasLoginForm = true
								break
							}
						}
					}
				}
			}
			if t.Data == "input" {
				for _, attr := range t.Attr {
					if attr.Key == "type" {
						if strings.Contains(strings.ToLower(attr.Val), "password") {
							numOfPasswordInputs++
						}
					}
				}
			}
		}
	}
	h.result.HasLoginForm = hasLoginForm
	return h
}
