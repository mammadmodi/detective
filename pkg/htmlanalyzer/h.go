package htmlanalyzer

import (
	"context"
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

type HeadingsCount struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

type LinksCount struct {
	Internal int `json:"internal"`
	External int `json:"external"`
	Pointer  int `json:"pointer"`
}

type Result struct {
	HTMLVersion            string         `json:"html_version"`
	PageTitle              string         `json:"page_title"`
	HeadingsCount          *HeadingsCount `json:"headings_count"`
	LinksCount             *LinksCount    `json:"links_count"`
	InaccessibleLinksCount int            `json:"inaccessible_links_count"`
	HasLoginForm           bool           `json:"has_login_form"`
}

// HTMLAnalyzer is a struct by which you can parse an html string
// and retrieve useful information about that.
type HTMLAnalyzer struct {
	m             sync.Mutex
	httpClient    *http.Client
	internalLinks []*url.URL
	externalLinks []*url.URL
	pointerLinks  []string

	// HostURL is needed to check that the link is external or internal.
	HostURL      *url.URL
	Logger       *zap.Logger
	HtmlDocument string
}

// New creates an HTMLAnalyzer object for the entered html string.
func New(hostURL *url.URL, htmlDocument string, httpTimeout time.Duration, logger *zap.Logger) *HTMLAnalyzer {
	return &HTMLAnalyzer{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		internalLinks: []*url.URL{},
		externalLinks: []*url.URL{},
		pointerLinks:  []string{},

		HostURL:      hostURL,
		Logger:       logger,
		HtmlDocument: htmlDocument,
	}
}

// Analyze analyzes the html document and returns a Result object.
func (h *HTMLAnalyzer) Analyze(ctx context.Context) (Result, error) {
	// TODO handle errors.
	htmlVersion, _ := h.getHTMLVersion()
	pageTitle, _ := h.getPageTitle()
	headingsCount, _ := h.getHeadingsCount()
	linksCount, _ := h.getLinksCount()
	inaccessibleLinksCount, _ := h.getInaccessibleLinksCount(ctx)
	hasLoginForm, _ := h.hasLoginForm()
	return Result{
		HTMLVersion:            htmlVersion,
		PageTitle:              pageTitle,
		HeadingsCount:          headingsCount,
		LinksCount:             linksCount,
		InaccessibleLinksCount: inaccessibleLinksCount,
		HasLoginForm:           hasLoginForm,
	}, nil
}

func (h *HTMLAnalyzer) getHTMLVersion() (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.HtmlDocument))
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return "Unknown HTML Version", nil
			}
			return "Unknown HTML Version", fmt.Errorf("error while tokenizing HTML: %v", err)
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
					return d.version, nil
				}
			}
			return "unknown", nil
		}
	}
}

func (h *HTMLAnalyzer) getPageTitle() (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.HtmlDocument))
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return "Unknown Page Title", nil
			}
			return "Unknown Page Title", fmt.Errorf("error while tokenizing HTML: %v", err)
		}

		td := tokenizer.Token().Data
		if tt == html.StartTagToken && td == "title" {
			tt = tokenizer.Next()
			if tt == html.TextToken {
				return tokenizer.Token().Data, nil
			}
			return "Empty Page Title", nil
		}
	}
}

func (h *HTMLAnalyzer) getHeadingsCount() (*HeadingsCount, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.HtmlDocument))
	headings := &HeadingsCount{}
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return headings, nil
			}
			return headings, fmt.Errorf("error while tokenizing HTML: %v", err)
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
}

func (h *HTMLAnalyzer) getLinksCount() (*LinksCount, error) {
	if err := h.parseAndSetLinks(); err != nil {
		return nil, err
	}

	return &LinksCount{
		Internal: len(h.internalLinks),
		External: len(h.externalLinks),
		Pointer:  len(h.pointerLinks),
	}, nil
}

func (h *HTMLAnalyzer) parseAndSetLinks() error {
	h.internalLinks = []*url.URL{}
	h.externalLinks = []*url.URL{}
	h.pointerLinks = []string{}
	tokenizer := html.NewTokenizer(strings.NewReader(h.HtmlDocument))
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error while tokenizing HTML: %v", err)
		}

		t := tokenizer.Token()
		if tt == html.StartTagToken && t.Data == "a" {
			for _, a := range t.Attr {
				if a.Key == "href" {
					// Empty href.
					if len(a.Val) == 0 {
						h.Logger.With(zap.String("href", a.Val)).Debug("url ignored because it was empty")
						continue
					}
					// Pointer links.
					if string(a.Val[0]) == "#" {
						h.Logger.With(zap.String("href", a.Val)).Debug("url ignored because it was a pointer")
						h.pointerLinks = append(h.pointerLinks, a.Val)
						continue
					}

					u, err := url.Parse(strings.TrimSpace(a.Val))
					if err != nil {
						h.Logger.With(zap.String("href", a.Val)).Debug("url ignored, could not parse url")
						continue
					}

					if u.Scheme != "" && !strings.Contains(u.Scheme, "http") {
						h.Logger.With(zap.String("href", a.Val), zap.String("scheme", u.Scheme)).
							Debug("url ignored, bad scheme")
						continue
					}

					if h.isInternalLink(u) {
						u = u.ResolveReference(h.HostURL)
						h.internalLinks = append(h.internalLinks, u)
						h.Logger.With(zap.String("url", u.String())).Info("marked as internal")
					} else {
						h.externalLinks = append(h.externalLinks, u)
						h.Logger.With(zap.String("url", u.String())).Info("marked as external")
					}
				}
			}
		}
	}
}

func (h *HTMLAnalyzer) isInternalLink(u *url.URL) bool {
	return u.Host == "" || strings.Contains(strings.ToLower(u.Host), h.HostURL.Host)
}

func (h *HTMLAnalyzer) getInaccessibleLinksCount(ctx context.Context) (int, error) {
	links := h.getNonPointerLinks()

	var inaccessibleLinksCount int
	inc := func() {
		h.m.Lock()
		inaccessibleLinksCount++
		h.m.Unlock()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(links))
	for _, u := range links {
		u := u
		go func() {
			defer wg.Done()
			if !h.isAccessibleURL(ctx, u) {
				h.Logger.With(zap.String("url", u.String())).Debug("url is not accessible")
				inc()
				return
			}
			h.Logger.With(zap.String("url", u.String())).Debug("url is accessible")
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		h.Logger.Debug("all go routines finished successfully")
		done <- struct{}{}
	}()

	select {
	case <-done:
		return inaccessibleLinksCount, nil
	case <-ctx.Done():
		return inaccessibleLinksCount, fmt.Errorf("process stopped due to context got done")
	}
}

func (h *HTMLAnalyzer) getNonPointerLinks() []*url.URL {
	var nonPointerLinks []*url.URL
	nonPointerLinks = append(nonPointerLinks, h.externalLinks...)
	nonPointerLinks = append(nonPointerLinks, h.internalLinks...)
	return nonPointerLinks
}

func (h *HTMLAnalyzer) isAccessibleURL(ctx context.Context, u *url.URL) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		h.Logger.Error("could not create request")
		return false
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.Logger.Error("could not perform request")
		return false
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	h.Logger.Error("response code is not 200")
	return false
}

func (h *HTMLAnalyzer) hasLoginForm() (bool, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.HtmlDocument))
	// If the html has a form which has one of the following keywords in it's identity attributes
	// we can be sure that page has login form.
	loginFormKeywords := []string{"login", "signin", "sign_in"}
	// If the html has password inputs only once we can say that it's a login form.
	// If number of password inputs is more than 1 so it's a sign up page or a reset password.
	var numOfPasswordInputs uint8
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return numOfPasswordInputs == 1, nil
			}
			return false, fmt.Errorf("error while tokenizing HTML: %v", err)
		}
		t := tokenizer.Token()
		if tt == html.StartTagToken {
			if t.Data == "form" {
				for _, a := range t.Attr {
					if a.Key == "name" || a.Key == "id" || a.Key == "action" {
						for _, k := range loginFormKeywords {
							if strings.Contains(strings.ToLower(a.Val), k) {
								return true, nil
							}
						}
					}
				}
			}
			if t.Data == "input" {
				for _, a := range t.Attr {
					if a.Key == "type" {
						if strings.Contains(strings.ToLower(a.Val), "password") {
							numOfPasswordInputs++
						}
					}
				}
			}
		}
	}
}
