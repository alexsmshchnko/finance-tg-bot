package entity

type Report struct {
	RepName  string
	RepParms map[string]string
}

type ReportResult struct {
	Name string
	Val  int
}
