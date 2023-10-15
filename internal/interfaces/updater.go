package interfaces

type Updater interface {
	Tick() (bool, error)
}
