package uid

import gonanoid "github.com/matoous/go-nanoid/v2"

func Generate() (string, error) {
	return gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 12)
}
