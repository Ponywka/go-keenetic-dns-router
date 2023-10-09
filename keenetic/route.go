package keenetic

type Route struct {
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Interface string `json:"interface"`
	Host      string `json:"host"`
	Gateway   string `json:"gateway"`
	No        bool   `json:"no"`
	Auto      bool   `json:"auto"`
	Reject    bool   `json:"reject"`
}
