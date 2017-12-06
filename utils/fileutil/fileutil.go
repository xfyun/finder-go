package fileutil

import "os"

func ExistPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if (err != nil) {
		if (os.IsNotExist(err)) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func GetSystemSeparator() string {
	s := "/"

	if (os.IsPathSeparator('\\')) {
		s = "\\"
	}

	return s
}
