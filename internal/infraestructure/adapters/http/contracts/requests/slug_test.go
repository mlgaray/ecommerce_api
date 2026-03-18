package requests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

func TestValidateSlug(t *testing.T) {
	t.Run("when slug is valid then returns nil", func(t *testing.T) {
		validSlugs := []string{
			"my-store",
			"abc",
			"store-name-123",
			"a1b",
			"test",
			"my-awesome-store",
		}
		for _, slug := range validSlugs {
			err := ValidateSlug(slug)
			assert.NoError(t, err, "slug %q should be valid", slug)
		}
	})

	t.Run("when slug is too short then returns error", func(t *testing.T) {
		shortSlugs := []string{"", "a", "ab"}
		for _, slug := range shortSlugs {
			err := ValidateSlug(slug)
			assert.Error(t, err, "slug %q should be invalid", slug)
			var badReq *httpErrors.BadRequestError
			assert.ErrorAs(t, err, &badReq)
			assert.Equal(t, domainErrors.InvalidSlugFormat, badReq.Message)
		}
	})

	t.Run("when slug is too long then returns error", func(t *testing.T) {
		// 64 characters exceeds max of 63
		longSlug := strings.Repeat("a", 64)
		err := ValidateSlug(longSlug)
		assert.Error(t, err)
		var badReq *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReq)
		assert.Equal(t, domainErrors.InvalidSlugFormat, badReq.Message)
	})

	t.Run("when slug is exactly min length then returns nil", func(t *testing.T) {
		err := ValidateSlug("abc")
		assert.NoError(t, err)
	})

	t.Run("when slug is exactly max length then returns nil", func(t *testing.T) {
		maxSlug := strings.Repeat("a", 63)
		err := ValidateSlug(maxSlug)
		assert.NoError(t, err)
	})

	t.Run("when slug has uppercase letters then returns error", func(t *testing.T) {
		err := ValidateSlug("My-Store")
		assert.Error(t, err)
		var badReq *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReq)
	})

	t.Run("when slug starts with hyphen then returns error", func(t *testing.T) {
		err := ValidateSlug("-my-store")
		assert.Error(t, err)
	})

	t.Run("when slug ends with hyphen then returns error", func(t *testing.T) {
		err := ValidateSlug("my-store-")
		assert.Error(t, err)
	})

	t.Run("when slug has consecutive hyphens then returns error", func(t *testing.T) {
		err := ValidateSlug("my--store")
		assert.Error(t, err)
	})

	t.Run("when slug has special characters then returns error", func(t *testing.T) {
		invalidSlugs := []string{
			"my_store",
			"my store",
			"my.store",
			"my@store",
			"my!store",
		}
		for _, slug := range invalidSlugs {
			err := ValidateSlug(slug)
			assert.Error(t, err, "slug %q should be invalid", slug)
		}
	})
}
