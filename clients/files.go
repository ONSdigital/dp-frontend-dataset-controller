package clients

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
)

// FilesAPIClient is an interface with methods required for getting metadata from Files API
type FilesAPIClient interface {
	GetFile(ctx context.Context, path string, authToken string) (files.FileMetaData, error)
}
