package dto

// DataTableRequest menampung parameter yang dikirim oleh DataTables JS
type DataTableRequest struct {
	Draw   int    `form:"draw"`
	Start  int    `form:"start"`
	Length int    `form:"length"`
	Search string `form:"search[value]"` // Search global
}

// DataTableResponse adalah format JSON yang diharapkan DataTables
type DataTableResponse struct {
	Draw            int         `json:"draw"`
	RecordsTotal    int64       `json:"recordsTotal"`
	RecordsFiltered int64       `json:"recordsFiltered"`
	Data            interface{} `json:"data"`
	Error           string      `json:"error,omitempty"`
}