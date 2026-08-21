package utils

import (
	"hash/fnv"
	"io"
	"math/rand"
	"uuid"
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
	var id uuid.UUID
	if _, err := io.ReadFull(rng, id[:]); err != nil {
		return "", err
	}
	id[6] = (id[6] & 0x0f) | 0x40 // Version 4
	id[8] = (id[8] & 0x3f) | 0x80 // Variant 10
	return id.String(), nil
}
