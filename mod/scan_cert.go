package mod

import (
	"crypto/tls"
	"os"
	"path/filepath"
)

func ScanCertificates(certFilePath string) (keyFilePath string) {
	parent := filepath.Dir(certFilePath)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}

	var keyFile string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(parent, entry.Name())
		_, err := tls.LoadX509KeyPair(certFilePath, fullPath)
		if err != nil {
			continue
		} else {
			keyFile = fullPath
			break
		}
	}

	return keyFile
}
