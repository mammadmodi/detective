package htmlanalysis

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSetGlobalHTTPClient(t *testing.T) {
	httpClient := &http.Client{}
	SetGlobalHTTPClient(httpClient)
	assert.Equal(t, httpClient, globalHTTPClient)
}

func TestSetGlobalLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	SetGlobalLogger(logger)
	assert.Equal(t, logger, globalLogger)
}

func TestNewHTMLAnalyzer(t *testing.T) {
	htmlDoc := "htmlDoc"
	hostURL := &url.URL{}
	ha := NewHTMLAnalyzer(htmlDoc, hostURL)
	assert.Equal(t, htmlDoc, ha.htmlDoc)
	assert.Equal(t, hostURL, ha.hostURL)
	assert.NotNil(t, ha.internalLinks, ha.externalLinks, ha.result)
	assert.False(t, ha.linksAreParsed)
}

func TestHTMLAnalyzer_GetHTMLVersion(t *testing.T) {
	testCases := []struct {
		expectedVersion string
		htmlDoc         string
	}{
		{
			expectedVersion: "HTML 5",
			htmlDoc:         `<!DOCTYPE HTML>`,
		},
		{
			expectedVersion: "HTML 4.01 Strict",
			htmlDoc:         `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN">`,
		},
		{
			expectedVersion: "HTML 4.01 Transitional",
			htmlDoc:         `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">`,
		},
		{
			expectedVersion: "HTML 4.01 Frameset",
			htmlDoc:         `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Frameset//EN">`,
		},
		{
			expectedVersion: "XHTML 1.0 Strict",
			htmlDoc:         `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN">`,
		},
		{
			expectedVersion: "XHTML 1.0 Transitional",
			htmlDoc:         `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN">`,
		},
		{
			expectedVersion: "XHTML 1.0 Frameset",
			htmlDoc:         `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN">`,
		},
		{
			expectedVersion: "XHTML 1.1",
			htmlDoc:         `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN">`,
		},
	}
	for _, tc := range testCases {
		t.Run("test html version "+tc.expectedVersion, func(t *testing.T) {
			a := NewHTMLAnalyzer(tc.htmlDoc, nil)
			version := a.GetHTMLVersion()
			assert.Equal(t, tc.expectedVersion, version)
		})
	}
}

func TestHTMLAnalyzer_GetPageTitle(t *testing.T) {
	testCases := []struct {
		name          string
		htmlDoc       string
		expectedTitle string
	}{
		{
			name:          "title tag is filled",
			htmlDoc:       `<title>Test Title</title>`,
			expectedTitle: "Test Title",
		},
		{
			name:          "title tag has no value",
			htmlDoc:       `<title></title>`,
			expectedTitle: "Empty Page Title",
		},
		{
			name:          "title tag is not exist",
			htmlDoc:       `<p>simple paragraph</p>`,
			expectedTitle: "Empty Page Title",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewHTMLAnalyzer(tc.htmlDoc, nil)
			pt := a.GetPageTitle()
			assert.Equal(t, tc.expectedTitle, pt)
		})
	}
}

func TestHTMLAnalyzer_GetHeadingsCount(t *testing.T) {
	htmlDoc, expectedHeadingsCount := generateTestHTMLDocWithHeadings()
	a := NewHTMLAnalyzer(htmlDoc, nil)
	hc := a.GetHeadingsCount()
	assert.Equal(t, expectedHeadingsCount, hc)
}

func TestHTMLAnalyzer_GetLinksCount(t *testing.T) {
	htmlDoc, linksCount, inaccessibleLinksCount, hostURL, shutdown := generateTestHTMLWithRealLinks()
	defer shutdown()
	a := NewHTMLAnalyzer(htmlDoc, hostURL)
	actualLinksCount := a.GetLinksCount()
	actualInaccessibleLinksCount := a.GetInaccessibleLinksCount(context.Background())
	assert.Equal(t, linksCount, actualLinksCount)
	assert.Equal(t, inaccessibleLinksCount, actualInaccessibleLinksCount)
}

func TestHTMLAnalyzer_HasLoginForm(t *testing.T) {
	testCases := []struct {
		name         string
		htmlDoc      string
		hasLoginForm bool
	}{
		{
			name:         "has no login form",
			htmlDoc:      `<form id="register">Register</form>`,
			hasLoginForm: false,
		},
		{
			name:         "has login form with an explicit id",
			htmlDoc:      `<form id="login">Register</form>`,
			hasLoginForm: true,
		},
		{
			name:         "has login implicit form",
			htmlDoc:      `<div>Login<input type="password">enter password</input></form>`,
			hasLoginForm: true,
		},
		{
			name:         "has login implicit form",
			htmlDoc:      `<div>signup<input type="password">enter pass</input><input type="password">re enter pass</input></form>`,
			hasLoginForm: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewHTMLAnalyzer(tc.htmlDoc, nil)
			ok := a.HasLoginForm()
			assert.Equal(t, tc.hasLoginForm, ok)
		})
	}
}

// getTestHTMLDocWithHeadings generates an html document which contains a lot of headings with different levels.
func generateTestHTMLDocWithHeadings() (htmlDoc string, expectedHeadings *HeadingsCount) {
	htmlDoc = `
<h1>H1</h1><h1>H1</h1><h1>H1</h1>
<h2>H2</h2><h2>H2</h2>
<h3>H3</h3><h3>H3</h3><h3>H3</h3><h3>H3</h3>
<h4>H4</h4><h4>H4</h4>
<h5>H5</h5><h5>H5</h5><h5>H5</h5><h5>H5</h5>
<h6>H6</h6><h6>H6</h6>
`
	expectedHeadings = &HeadingsCount{
		H1: 3,
		H2: 2,
		H3: 4,
		H4: 2,
		H5: 4,
		H6: 2,
	}
	return
}

// generateTestHTMLWithRealLinks sets up three different server,first is an internal, second one is an external
// and last one is an unavailable external server.
// after that it generates an html document which contains links to those servers which is useful for test purposes.
// shutdown is a function which stops test servers.
func generateTestHTMLWithRealLinks() (
	htmlDoc string,
	linksCount *LinksCount,
	inaccessibleLinksCount int,
	hostURL *url.URL,
	shutdown func(),
) {
	// Setting up test servers.
	availableIntServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))
	availableExtServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusOK)
	}))
	unavailableServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusServiceUnavailable)
	}))
	availableIntServerURL, _ := url.Parse(availableIntServer.URL)
	availableExtServerURL, _ := url.Parse(availableExtServer.URL)
	unavailableServerURL, _ := url.Parse(unavailableServer.URL)

	shutdown = func() {
		defer availableIntServer.Close()
		defer availableExtServer.Close()
		defer unavailableServer.Close()
	}

	// Create template html document string.
	format := "<a href=\"%s\">%s</a>\n"
	var testHrefs = []struct {
		name       string
		href       string
		internal   bool
		valid      bool
		accessible bool
	}{
		{name: "empty link", href: "", internal: true, valid: false, accessible: true},
		{name: "pointer link", href: "#top", internal: true, valid: false, accessible: true},
		{name: "internal relative link", href: "contact-us", internal: true, valid: true, accessible: true},
		{name: "internal relative link", href: "login.php", internal: true, valid: true, accessible: true},
		{name: "internal link", href: availableIntServerURL.String(), internal: true, valid: true, accessible: true},
		{name: "external link", href: availableExtServerURL.String(), internal: false, valid: true, accessible: true},
		{name: "unavailable link", href: unavailableServerURL.String(), internal: false, valid: true, accessible: false},
	}

	linksCount = &LinksCount{}
	for _, th := range testHrefs {
		htmlDoc = htmlDoc + fmt.Sprintf(format, th.href, th.name)
		if th.valid {
			if !th.accessible {
				inaccessibleLinksCount++
			}
			if th.internal {
				linksCount.Internal++
			} else {
				linksCount.External++
			}
		}
	}
	hostURL, _ = url.ParseRequestURI(availableIntServerURL.String())

	return
}
