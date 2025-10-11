package testcontainers

import (
	"errors"
	"net"
)

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = l.Close()
		if err != nil {
			panic(err)
		}
	}()

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("failed to get TCP address")
	}

	return tcpAddr.Port, nil
}
