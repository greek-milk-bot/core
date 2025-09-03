package core

type Resource struct {
	PluginID string `json:"id"`
	Scheme   string `json:"scheme"`
	Body     string `json:"body"`
}
