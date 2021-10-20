package sshserver

type exitSignalPayload struct {
	Signal       string
	CoreDumped   bool
	ErrorMessage string
	LanguageTag  string
}
