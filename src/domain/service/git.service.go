package service

type GitService interface {
	CloneRepository(url string, branch string, path string) error
	Checkout(path string, branch string) error
	Branch(path string, branch string) error
	Commit(path string, message string) error
	Push(path string, branch string) error
	Pull(path string, branch string) error
}
