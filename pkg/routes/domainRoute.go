package routes

type DomainRoute struct {
	Domain    string
	Comment   string
	Interface string
	Gateway   string
	Auto      bool
	Reject    bool
}
