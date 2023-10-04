package service

type DirectoryService interface {
	CopyFile(src string, dest string) error
	CopyDirectory(src string, dest string) error
	CreateDirectory(path string) error
	RemoveDirectory(path string) error
	DirectoryExists(path string) (bool, error)
	ApplyTemplateRecursively(path string, values interface{}) error
	ApplyTemplate(path string, values interface{}) error
	LoadTemplate(path string, values interface{}, html bool) ([]byte, error)
	RenameFile(oldPath, newPath string) error
	DeleteFile(path string) error
	VerifyOrInsertLineInFile(path string, line string) error
}
