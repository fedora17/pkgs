package utils

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"syscall"
)

func ReadPassword() (string, error) {
	fmt.Print("Enter Password: ")
	var fd int
	if terminal.IsTerminal(syscall.Stdin) {
		fd = syscall.Stdin
	} else {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return "", err
		}
		defer tty.Close()
		fd = int(tty.Fd())
	}

	pass, err := terminal.ReadPassword(fd)
	fmt.Println()
	return string(pass), err
}
