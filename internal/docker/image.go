package docker

import (
	"errors"
	"strings"

	"github.com/docker/distribution/reference"
)

func getCanonicalImageName(image string) (string, error) {
	_, err := reference.ParseNamed(image)
	if err != nil {
		if errors.Is(err, reference.ErrNameNotCanonical) {
			if !strings.Contains(image, "/") {
				image = "docker.io/library/" + image
			} else {
				image = "docker.io/" + image
			}
		} else {
			return "", err
		}
	}
	return image, nil
}
