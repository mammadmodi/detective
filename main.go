package main

import (
	"net/url"
)

func main() {
	u1, _ := url.Parse("/path")
	u2 := url.URL{
		Path: "https://www.google.com/path",
	}
	println(u1.String())
	println(u1.ResolveReference(&u2).String())
	NewAnalyzer("asdsad").
}

type Report struct {
	headers int
	links   int
}

type Analyzer struct {
	Doctype string
	Report  *Report
}

func NewAnalyzer(Doctype string) *Analyzer {
	return &Analyzer{
		Doctype: Doctype,
		Report:  &Report{},
	}
}

func (a *Analyzer) GetReport() *Report {
	return a.setLinks().setHeaders().Report
}

func (a *Analyzer) setHeaders() *Analyzer {
	// TODO logic
	a.Report.headers = 5
	return a
}

func (a *Analyzer) setLinks() *Analyzer {
	// TODO logic
	a.Report.links = 5
	return a
}
