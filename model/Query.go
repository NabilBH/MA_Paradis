package model

type Query struct {
	Id            string   `"yaml" : "id"`
	TTL           int      `"yaml": "ttl"`
	Data          string   `"yaml": "data"`
	Type          int      `"yaml": "type"`
	SourceAddress string   `"yaml" : "sourceaddress"`
	Path          []string `"yaml": "path"`
}
