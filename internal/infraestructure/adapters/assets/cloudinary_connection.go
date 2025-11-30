package assets

import (
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

// CloudinaryConnection defines the interface for Cloudinary connection.
type CloudinaryConnection interface {
	Connect() *cloudinary.Cloudinary
}

type cloudinaryConnection struct{}

// NewCloudinaryConnection creates a new Cloudinary connection factory.
func NewCloudinaryConnection() *cloudinaryConnection {
	return &cloudinaryConnection{}
}

// Connect establishes a connection to Cloudinary using environment variables.
func (c *cloudinaryConnection) Connect() *cloudinary.Cloudinary {
	cloud := os.Getenv("CLOUDINARY_CLOUD")
	key := os.Getenv("CLOUDINARY_KEY")
	secret := os.Getenv("CLOUDINARY_SECRET")

	cld, err := cloudinary.NewFromParams(cloud, key, secret)
	if err != nil {
		log.Fatalf("Failed to connect to Cloudinary: %v", err)
	}
	cld.Config.URL.Secure = true

	log.Println("Successfully connected to Cloudinary!")
	return cld
}
