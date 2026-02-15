//go:build darwin
// +build darwin

package comm

func DefaultEtcHosts() (string, error) {
	return "/private/etc/hosts", nil
}
