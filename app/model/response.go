package model

type WebResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code,omitempty"`
	Message   string            `json:"message"`
	Data      any               `json:"data,omitempty"`
	Meta      any               `json:"meta,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type ListQuery struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}