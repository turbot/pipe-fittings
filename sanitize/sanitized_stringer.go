package sanitize

type SanitizedStringer interface {
	String(sanitizer *Sanitizer, opts RenderOptions) string
}
