package cpit

// BaseModel is a struct that contains the common fields for all models
type BaseModel struct {
	ID         string `json:"_id"`
	State      int    `json:"_state"`
	Modified   int    `json:"_modified"`
	ModifiedBy string `json:"_mby"`
	Created    int    `json:"_created"`
	CreatedBy  string `json:"_cby"`
}

// File is a struct that contains the fields for the file model
type File struct {
	BaseModel
	Hash        string   `json:"_hash"`
	Thumbhash   string   `json:"thumbhash"`
	Path        string   `json:"path"`
	Title       string   `json:"title"`
	Mime        string   `json:"mime"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Tags        []any    `json:"tags"`
	Size        int      `json:"size"`
	Colors      []string `json:"colors"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Folder      string   `json:"folder"`
}

// PaginatedResp is a struct that contains the fields for a paginated response
type PaginatedResp[T any] struct {
	Data []T `json:"data"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type UpsertData struct {
	Data interface{} `json:"data"`
}
