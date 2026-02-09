package contracts

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"

	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

func TestCategoryUpdateRequest_Validate(t *testing.T) {
	t.Run("when request is valid with existing image then returns no error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when request is valid with new image then returns no error", func(t *testing.T) {
		// Arrange
		// Create a valid PNG image header (89 50 4E 47 0D 0A 1A 0A)
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		imageContent := make([]byte, len(pngHeader)+100)
		copy(imageContent, pngHeader)

		fileHeader := createTestFileHeader("image.png", imageContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image:       nil,
			},
			NewImage: fileHeader,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when name is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "category_name_is_required", badRequestErr.Message)
	})

	t.Run("when name is only whitespace then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "   ",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "category_name_is_required", badRequestErr.Message)
	})

	t.Run("when description is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "category_description_is_required", badRequestErr.Message)
	})

	t.Run("when description is only whitespace then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "   ",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "category_description_is_required", badRequestErr.Message)
	})

	t.Run("when no existing image and no new image then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image:       nil,
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "category_image_required", badRequestErr.Message)
	})

	t.Run("when existing image has invalid ID then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  0, // Invalid ID
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "existing_image_must_have_valid_id", badRequestErr.Message)
	})

	t.Run("when existing image has negative ID then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  -1, // Negative ID
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "existing_image_must_have_valid_id", badRequestErr.Message)
	})

	t.Run("when existing image has empty URL then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "existing_image_must_have_url", badRequestErr.Message)
	})

	t.Run("when existing image has whitespace URL then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "   ",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "existing_image_must_have_url", badRequestErr.Message)
	})

	t.Run("when new image exceeds 3MB then returns bad request error", func(t *testing.T) {
		// Arrange
		// Create file header with size > 3MB
		largeContent := make([]byte, 4*1024*1024) // 4MB
		fileHeader := createTestFileHeader("large_image.jpg", largeContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image:       nil,
			},
			NewImage: fileHeader,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "image_size_too_large_max_3mb", badRequestErr.Message)
	})

	t.Run("when new image has invalid MIME type then returns bad request error", func(t *testing.T) {
		// Arrange
		// Create a text file content (not an image)
		textContent := []byte("This is a text file, not an image")
		fileHeader := createTestFileHeader("document.txt", textContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image:       nil,
			},
			NewImage: fileHeader,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "invalid_image_type_only_jpeg_png_allowed", badRequestErr.Message)
	})

	t.Run("when new image is valid JPEG then returns no error", func(t *testing.T) {
		// Arrange
		// Create a valid JPEG header (FF D8 FF)
		jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		imageContent := make([]byte, len(jpegHeader)+100)
		copy(imageContent, jpegHeader)
		fileHeader := createTestFileHeader("image.jpg", imageContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image:       nil,
			},
			NewImage: fileHeader,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when new image is valid PNG then returns no error", func(t *testing.T) {
		// Arrange
		// Create a valid PNG header
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		imageContent := make([]byte, len(pngHeader)+100)
		copy(imageContent, pngHeader)
		fileHeader := createTestFileHeader("image.png", imageContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image:       nil,
			},
			NewImage: fileHeader,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when replacing existing image with new image then only validates new image", func(t *testing.T) {
		// Arrange
		// Create a valid PNG image
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		imageContent := make([]byte, len(pngHeader)+100)
		copy(imageContent, pngHeader)
		fileHeader := createTestFileHeader("new_image.png", imageContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/old_image.jpg",
				},
			},
			NewImage: fileHeader,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when both name and description are empty then returns name error first", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "",
				Description: "",
				Image: &CategoryImageRequest{
					ID:  1,
					URL: "https://cloudinary.com/image.jpg",
				},
			},
			NewImage: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "category_name_is_required", badRequestErr.Message)
	})
}

func TestCategoryUpdateRequest_ToImageBuffer(t *testing.T) {
	t.Run("when new image is provided then returns image bytes", func(t *testing.T) {
		// Arrange
		imageContent := []byte("test image content")
		fileHeader := createTestFileHeader("image.jpg", imageContent)

		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
			},
			NewImage: fileHeader,
		}

		// Act
		buffer, err := request.ToImageBuffer()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, buffer)
		assert.Equal(t, imageContent, buffer)
	})

	t.Run("when no new image then returns nil", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Electronics",
				Description: "Electronic products",
			},
			NewImage: nil,
		}

		// Act
		buffer, err := request.ToImageBuffer()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, buffer)
	})
}

// createTestFileHeader creates a multipart.FileHeader for testing purposes
func createTestFileHeader(filename string, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("image", filename)
	_, _ = part.Write(content)
	writer.Close()

	// Parse the multipart form to get the FileHeader
	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(10 << 20) // 10 MB max memory

	if files, ok := form.File["image"]; ok && len(files) > 0 {
		return files[0]
	}

	return nil
}
