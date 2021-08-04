package template

type DatasetSnapshot struct {
	Name string `json:"name"`
}
type DatasetTemplate struct {
	Snapshots []DatasetSnapshot `json:"snapshots"`
}
