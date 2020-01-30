package main

type Config struct {
	CredentialJson Credential   `yaml:"firebase"`
	Collection     []Collection `yaml:"collections"`
}
type Collection struct {
	Path              string   `yaml:"path"`
	FTS               bool     `yaml:"fts"`
	IsGroupCollection bool     `yaml:"group-collections"`
	FtsField          []Field  `yaml:"fields"`
	Filter            []Filter `yaml:"filter"`
}
type Credential struct {
	Path string `yaml:"cred"`
}
type Filter struct {
	Field   string      `yaml:"field"`
	Compare string      `yaml:"compare"`
	Value   interface{} `yaml:"value"`
	Raw     bool
}
type Field struct {
	Name string `yaml:"field"`
}
type SqliteMaster struct {
	Type     string
	TblName  string
	RootPage string
	Name     string
	Sql      string
}
type DataParser struct {
	Field   []string `json:"fields"`
	Limit   uint64   `json:"limit"`
	Offset  uint64   `json:"offset"`
	Sort    []string `json:"sorts"`
	Filter  []Filter `json:"filters"`
	GroupBy []string `json:"groupbys"`
}
type AggParser struct {
	Field   string   `json:"fields"`
	AggFuns string   `json:"aggfuns"`
	Limit   uint64   `json:"limit"`
	Offset  uint64   `json:"offset"`
	Sort    []string `json:"sorts"`
	GroupBy string   `json:"groupbys"`
	Filter  []Filter `json:"filters"`
}

type Return interface{}
type Heartbeat struct {
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Mem        uint64 `json:"memory"`
}

func (d DataParser) HasFields() bool {
	return len(d.Field) > 1
}
func (d DataParser) HasFilters() bool {
	return len(d.Filter) > 0
}
func (d DataParser) HasGroupBys() bool {
	return len(d.GroupBy) > 0
}
func (d DataParser) HasSorts() bool {
	return len(d.Sort) > 0
}
