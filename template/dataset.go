package template

type DatasetSnapshot struct {
	Name string
}
type DatasetTemplate struct {
	Snapshots []DatasetSnapshot `json:"snapshots"`
}
