package utils

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"

	"github.com/pquerna/otp"
)

func TotpKeyToImage(key *otp.Key) (string, error) {
	keyImg, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}

	var imgBuffer bytes.Buffer
	err = jpeg.Encode(&imgBuffer, keyImg, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(imgBuffer.Bytes())
	return encoded, nil
}
