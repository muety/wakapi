package utils

import (
	"hash/fnv"
	"io"
	"math/rand"

	"github.com/gofrs/uuid/v5"
)

func RandFromSeedString(seed string) (*rand.Rand, error) {
	hash := fnv.New64a()
	if _, err := io.WriteString(hash, seed); err != nil {
		return nil, err
	}
	return rand.New(rand.NewSource(int64(hash.Sum64()))), nil
}

func UUIDFromSeed(seed string) (string, error) {
	rng, err := RandFromSeedString(seed)
	if err != nil {
		return "", err
	}
	gen := uuid.NewGenWithOptions(uuid.WithRandomReader(rng))
	id, err := gen.NewV4()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
