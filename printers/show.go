package printers

// interface implemented by objectes ewhich support the `show` command for pretty/plain output format
type Showable interface {
	GetShowData() *TableRow
}

func IsShowable(value any) bool {
	return AsShowable(value) != nil
}

func AsShowable(value any) Showable {
	if s, ok := value.(Showable); ok {
		return s
	}
	// check the pointer
	pVal := &value
	if s, ok := any(pVal).(Showable); ok {
		return s
	}
	return nil
}

// interface implemented by objectes ewhich support the `list` command for pretty/plain output format
type Listable interface {
	GetListData() *TableRow
}

func AsListable(value any) Listable {
	if l, ok := value.(Listable); ok {
		return l
	}
	// check the pointer
	pVal := &value
	if l, ok := any(pVal).(Listable); ok {
		return l
	}
	return nil
}
