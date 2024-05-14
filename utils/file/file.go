package file

import "os"

func Write(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}
