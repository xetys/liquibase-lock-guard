package pkg

import "os"

func Getenv(name string, defaultValue string) string {
	value := os.Getenv(name)

	if len(value) == 0 {
		value = defaultValue
	}

	return value
}
