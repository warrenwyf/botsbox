package fetchers

type Fetcher interface {
	Fetch() (result *Result, err error)
	Hash() string // Used to compare two fetchers are equal or not
}

type Result struct {
	Hash    string
	Format  int
	Content interface{}
}

const (
	ResultFormat_Bytes = iota
	ResultFormat_Browser
)
