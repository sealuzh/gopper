package util

import (
	"fmt"
	"os/user"
	"path/filepath"
)

func AbsolutePath(p string) string {
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("No current user available: %v", err))
	}
	tilde := "~/"
	if len(p) > 2 && p[:2] == tilde {
		return filepath.Join(usr.HomeDir, p[2:])
	}
	return p
}
