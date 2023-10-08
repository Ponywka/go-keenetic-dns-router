package keenetic

// TODO: Type for summary
type InterfaceBase struct {
	Id            string      `json:"id"`
	Index         int         `json:"index"`
	InterfaceName string      `json:"interface-name"`
	Type          string      `json:"type"`
	Traits        []string    `json:"traits"`
	Link          string      `json:"link"`
	Summary       interface{} `json:"summary"`
}
