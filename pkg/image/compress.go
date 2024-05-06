package image

import (
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"strings"
)

func CompressImage(buffer []byte, quality int) ([]byte, string, error) {
	if quality <= 0 || quality > 100 {
		return nil, "", domain.ErrImageQuality
	}

	filename := strings.Replace(uuid.New().String(), "-", "", -1) + ".webp"

	converted, err := bimg.NewImage(buffer).Convert(bimg.WEBP)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "Unsupported image format"):
			return nil, "", domain.ErrImageFormat
		case strings.Contains(err.Error(), "Image buffer is empty"):
			return nil, "", domain.ErrImageIsEmpty
		}
		return nil, "", err
	}

	processed, err := bimg.NewImage(converted).Process(bimg.Options{Quality: quality})
	if err != nil {
		return nil, "", err
	}

	return processed, filename, nil
}
