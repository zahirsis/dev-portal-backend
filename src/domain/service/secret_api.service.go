package service

type SecretApiService interface {
	CreateBlank(location, path string) error
}
