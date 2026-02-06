package file

const (
	FolderUser         = "user"
	FolderMessageMedia = "message_media"
)

var (
	defaultAllowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tif", ".tiff", ".webp", ".heic", ".heif", ".raw"}
	defaultAllowedVideoExtensions = []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".3gp", ".m4v", ".mpeg", ".mpg", ".ogv"}
	defaultAllowedExtensions      = append(defaultAllowedImageExtensions, defaultAllowedVideoExtensions...)
)

const (
	defaultMaxSizeBytes      = 100 * 1024 * 1024 // 100 MB
	defaultMaxImageSizeBytes = 10 * 1024 * 1024  // 10 MB
	clientRootPath           = "resources/clients"
	temporaryFolderName      = "temporary"
)

type folderConfig struct {
	allowedExtensions []string
	maxSizeBytes      int64
}

var folderConfigurations = map[string]folderConfig{
	temporaryFolderName: {
		allowedExtensions: defaultAllowedExtensions,
		maxSizeBytes:      defaultMaxSizeBytes,
	},
	FolderUser: {
		allowedExtensions: defaultAllowedImageExtensions,
		maxSizeBytes:      defaultMaxImageSizeBytes,
	},
	FolderMessageMedia: {
		allowedExtensions: defaultAllowedExtensions,
		maxSizeBytes:      defaultMaxSizeBytes,
	},
}
