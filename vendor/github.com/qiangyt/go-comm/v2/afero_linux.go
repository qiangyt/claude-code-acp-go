//go:build linux
// +build linux

package comm

func DefaultEtcHosts() (string, error) {
	return "/etc/hosts", nil
}
