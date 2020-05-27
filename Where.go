package db

type Where struct {
	join    string // and or
	combine string // and or
	item    []string
}
