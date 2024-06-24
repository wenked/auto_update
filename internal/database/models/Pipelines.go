package models

type Pipeline struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UpdatePipeline struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
