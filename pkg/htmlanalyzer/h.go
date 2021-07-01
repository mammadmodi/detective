package htmlanalyzer

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Headings struct {
	H1, H2, H3, H4, H5, H6 uint
}

type Result struct {
	HTMLVersion       string   `json:"html_version"`
	PageTitle         string   `json:"page_title"`
	HeadingsCount     Headings `json:"headings"`
	InternalLinks     uint     `json:"internal_links"`
	ExternalLinks     uint     `json:"external_links"`
	InaccessibleLinks uint     `json:"inaccessible_links"`
	HasLoginForm      bool     `json:"has_login_form"`
}

// HTMLAnalyzer is a struct by which you can parse an html string
// and retrieve useful information about that.
type HTMLAnalyzer struct {
	// Host is needed to check that the link is external or internal.
	Host string

	htmlDocument  string
	internalLinks []*url.URL
	externalLinks []*url.URL
}

// New creates an HTMLAnalyzer object for the entered html string.
func New(host, htmlDocument string) *HTMLAnalyzer {
	return &HTMLAnalyzer{
		Host:         strings.ToLower(host),
		htmlDocument: htmlDocument,
	}
}

// Analyze analyzes the html document and returns a Result object.
func (h HTMLAnalyzer) Analyze() (Result, error) {
	// TODO handle errors.
	htmlVersion, _ := h.getHTMLVersion()
	pageTitle, _ := h.getPageTitle()
	headingsCount, _ := h.getHeadingsCount()
	intLinksCount, extLinksCount, _ := h.getLinksCount()
	return Result{
		HTMLVersion:       htmlVersion,
		PageTitle:         pageTitle,
		HeadingsCount:     headingsCount,
		InternalLinks:     uint(intLinksCount),
		ExternalLinks:     uint(extLinksCount),
		InaccessibleLinks: 0,
		HasLoginForm:      false,
	}, nil
}

func (h *HTMLAnalyzer) getHTMLVersion() (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDocument))
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
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDocument))
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

func (h *HTMLAnalyzer) getHeadingsCount() (Headings, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDocument))
	headings := Headings{}
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

func (h *HTMLAnalyzer) getLinksCount() (int, int, error) {
	if err := h.parseAndSetLinks(); err != nil {
		return 0, 0, err
	}

	return len(h.internalLinks), len(h.externalLinks), nil
}

func (h *HTMLAnalyzer) parseAndSetLinks() error {
	h.internalLinks = []*url.URL{}
	h.externalLinks = []*url.URL{}
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDocument))
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
					u, err := url.Parse(strings.TrimSpace(a.Val))
					if err != nil {
						continue
					}
					if u.Host == "" || strings.Contains(strings.ToLower(u.Host), h.Host) {
						h.internalLinks = append(h.internalLinks, u)
					} else {
						h.externalLinks = append(h.externalLinks, u)
					}
				}
			}
		}
	}
}
