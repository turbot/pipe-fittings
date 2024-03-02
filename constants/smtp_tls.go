package constants

const (
	SmtpTlsRequired = "required"
	SmtpTlsOff      = "off"
	SmtpTlsAuto     = "auto"
)

func IsValidSmtpTls(s string) bool {
	switch s {
	case SmtpTlsRequired, SmtpTlsOff, SmtpTlsAuto:
		return true
	default:
		return false
	}
}
