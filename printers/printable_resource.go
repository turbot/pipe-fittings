package printers

type Printer int

type PrintableResource[T any] interface {
	GetItems() []T
	GetTable() (Table, error)
}
