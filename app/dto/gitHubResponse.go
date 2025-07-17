package dto

type GitHubPRResponse struct {
	Title        string `json:"title"`
	Body         string `json:"body"`
	State        string `json:"state"` // open, closed, merged
	Merged       bool   `json:"merged"`
	ChangedFiles int    `json:"changed_files"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	Commits      int    `json:"commits"`
	Description  string `json:"description"`

	Base struct {
		Ref string `json:"ref"` // target branch (e.g. master/main)
	} `json:"base"`

	Head struct {
		Ref string `json:"ref"` // source branch (e.g. feature/product-page)
	} `json:"head"`
}
