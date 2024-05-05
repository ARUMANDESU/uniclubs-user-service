package image

import (
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestCompressImage(t *testing.T) {
	image, err := os.ReadFile("../../tests/image/test_image.jpg")
	require.NoError(t, err)
	quality := 70

	_, _, err = CompressImage(image, quality)

	require.NoError(t, err)
}

func TestCompressImage_ImageIsNotValid(t *testing.T) {
	tests := []struct {
		name        string
		image       []byte
		expectedErr error
	}{
		{
			name:        "empty image",
			image:       []byte{},
			expectedErr: domain.ErrImageIsEmpty,
		},
		{
			name:        "nil image",
			image:       nil,
			expectedErr: domain.ErrImageIsEmpty,
		},
		{
			name:        "not an image",
			image:       []byte("not an image"),
			expectedErr: domain.ErrImageFormat,
		},
	}
	quality := 70

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := CompressImage(tt.image, quality)

			require.ErrorIs(t, err, tt.expectedErr, "image: %v", tt.image)
		})
	}
}

func TestCompressImage_QualityIsNotValid(t *testing.T) {
	image, err := os.ReadFile("../../tests/image/test_image.jpg")
	require.NoError(t, err)

	tests := []int{-1, 0, 101, 1000}

	for _, quality := range tests {
		_, _, err = CompressImage(image, quality)

		require.ErrorIs(t, err, domain.ErrImageQuality)

	}
}
