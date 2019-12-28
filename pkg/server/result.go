package server

type PagedMentionList struct {
	Items []Mention
	Total int
	Next  string
}
