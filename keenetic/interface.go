package keenetic

type InterfaceBase struct {
	Id            string        `json:"id"`
	Index         float64       `json:"index"`
	InterfaceName string        `json:"interface-name"`
	Type          string        `json:"type"`
	Traits        []interface{} `json:"traits"`
	Link          string        `json:"link"`
	Summary       interface{}   `json:"summary"`
}
