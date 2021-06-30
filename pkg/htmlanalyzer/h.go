package htmlanalyzer

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Headings struct {
	H1, H2, H3, H4, H5, H6 uint
}

type Result struct {
	HTMLVersion       string    `json:"html_version"`
	PageTitle         string    `json:"page_title"`
	HeadingsCount     *Headings `json:"headings"`
	InternalLinks     uint      `json:"internal_links"`
	ExternalLinks     uint      `json:"external_links"`
	InaccessibleLinks uint      `json:"inaccessible_links"`
	HasLoginForm      bool      `json:"has_login_form"`
}

// HTMLAnalyzer is a struct by which you can parse an html string
// and retrieve useful information about that.
type HTMLAnalyzer struct {
	htmlDocument string
}

// New creates an HTMLAnalyzer object for the entered html string.
func New(htmlDocument string) *HTMLAnalyzer {
	return &HTMLAnalyzer{
		htmlDocument: htmlDocument,
	}
}

// Analyze analyzes the html document and returns a Result object.
func (h HTMLAnalyzer) Analyze() (Result, error) {
	// TODO handle errors.
	htmlVersion, _ := h.resolveHTMLVersion()
	return Result{
		HTMLVersion:       htmlVersion,
		PageTitle:         "",
		HeadingsCount:     nil,
		InternalLinks:     0,
		ExternalLinks:     0,
		InaccessibleLinks: 0,
		HasLoginForm:      false,
	}, nil
}

func (h *HTMLAnalyzer) resolveHTMLVersion() (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(h.htmlDocument))
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return "unknown_version", nil
			}
			return "unknown_version", fmt.Errorf("error while tokenizing HTML: %v", err)
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
