package server

type PagedMentionList struct {
	Items []Mention `json:"items"`
	Total int       `json:"total"`
	Next  string    `json:"next,omitempty"`
}
