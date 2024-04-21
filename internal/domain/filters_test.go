package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateMetadata(t *testing.T) {
	tests := []struct {
		totalRecords   int32
		page           int32
		pageSize       int32
		expectedResult Metadata
	}{
		{
			totalRecords: 5,
			page:         1,
			pageSize:     10,
			expectedResult: Metadata{
				CurrentPage:  1,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     1,
				TotalRecords: 5,
			},
		},
		{
			totalRecords:   0,
			page:           1,
			pageSize:       10,
			expectedResult: Metadata{},
		},
		{
			totalRecords: 20,
			page:         2,
			pageSize:     10,
			expectedResult: Metadata{
				CurrentPage:  2,
				PageSize:     10,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			metadata := CalculateMetadata(tt.totalRecords, tt.page, tt.pageSize)

			assert.Equal(t, tt.expectedResult, metadata)
		})
	}
}

func TestFilters_Limit(t *testing.T) {
	tests := []struct {
		filters       Filters
		expectedLimit int32
	}{
		{
			filters: Filters{
				Page:     1,
				PageSize: 10,
			},
			expectedLimit: 10,
		},
		{
			filters: Filters{
				Page:     1,
				PageSize: 1,
			},
			expectedLimit: 1,
		},
		{
			filters: Filters{
				Page:     1,
				PageSize: -1,
			},
			expectedLimit: -1,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.filters.Limit()

			assert.Equal(t, tt.expectedLimit, result)
		})
	}

}

func TestFilters_Offset(t *testing.T) {
	tests := []struct {
		filters        Filters
		expectedOffset int32
	}{
		{
			filters: Filters{
				Page:     1,
				PageSize: 10,
			},
			expectedOffset: 0,
		},
		{
			filters: Filters{
				Page:     2,
				PageSize: 20,
			},
			expectedOffset: 20,
		},
		{
			filters: Filters{
				Page:     5,
				PageSize: 30,
			},
			expectedOffset: 120,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.filters.Offset()

			assert.Equal(t, tt.expectedOffset, result)
		})
	}
}
