package apidadaptor

type ScopeSetResult struct {
	Name   string  `json:"name"`
	Scopes []Scope `json:"scopes"`
}
type Scope struct {
	Name string `json:"scope"`
}
