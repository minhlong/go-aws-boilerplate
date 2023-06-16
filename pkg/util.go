package util

import (
	"log"
	"os"
)

func MustGetEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("Env %s is missing", name)
	}
	return value
}
