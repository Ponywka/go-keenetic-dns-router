package keenetic

// TODO: Type for summary
type InterfaceBase struct {
	Id      string   `json:"id"`
	Index   int      `json:"index"`
	Name    string   `json:"interface-name"`
	Type    string   `json:"type"`
	Traits  []string `json:"traits"`
	Link    string   `json:"link"`
	Summary struct {
		Layer struct {
			Conf string `json:"conf"`
			Link string `json:"link"`
		} `json:"layer"`
	} `json:"summary"`
}
