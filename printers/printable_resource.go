package printers

type PrintableResource[T any] interface {
	GetItems() []T
	GetTable() (Table, error)
}
