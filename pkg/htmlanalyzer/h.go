package htmlanalyzer

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
	htmlStr string
}

// New creates an HTMLAnalyzer object for the entered html string.
func New(htmlStr string) *HTMLAnalyzer {
	return &HTMLAnalyzer{
		htmlStr: htmlStr,
	}
}

func (h HTMLAnalyzer) Analyze() (Result, error) {
	//TODO analyze the html.
	return Result{}, nil
}
