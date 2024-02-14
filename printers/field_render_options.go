package printers

type FieldRenderOptions struct {
	// a function implementing custom rendering logic to display the value
	renderValueFunc RenderFunc
	listOpts        listFieldRenderOptions
}

type listFieldRenderOptions struct {
	// a function implementing custom rendering logic to display the key AND value
	listKeyRenderFunc RenderFunc
	// is this the key field - if not it will be indented when rendering in a list
	isKey bool
}
