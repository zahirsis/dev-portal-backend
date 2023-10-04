package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"os"
	"strconv"
	"time"
)

type GitService string

const (
	GitBitbucket GitService = "bitbucket"
)

func GitServiceFromString(service string) GitService {
	switch service {
	case "bitbucket":
		return GitBitbucket
	default:
		return GitBitbucket
	}
}

type WikiService string

const (
	WikiConfluence WikiService = "confluence"
)

func WikiServiceFromString(service string) WikiService {
	switch service {
	case "confluence":
		return WikiConfluence
	default:
		return WikiConfluence
	}
}

type SecretService string

const (
	SecretVault SecretService = "vault"
)

func SecretServiceFromString(service string) SecretService {
	switch service {
	case "vault":
		return SecretVault
	default:
		return SecretVault
	}
}

type GitProtocol int

const (
	GitSSH GitProtocol = iota
	GitHTTPS
)

func GitProtocolFromString(protocol string) GitProtocol {
	switch protocol {
	case "ssh":
		return GitSSH
	case "https":
		return GitHTTPS
	default:
		return GitHTTPS
	}
}

type corsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type httpConfig struct {
	Path string
	Port string
}

type wsConfig struct {
	ReadBufferSize   int
	WriteBufferSize  int
	HandshakeTimeout time.Duration
}

type redisConfig struct {
	Addr     string
	Password string
	DB       int
}

type GitConfig struct {
	Host     string
	UserName string
	Token    string
	Project  string
	Protocol GitProtocol
}

type WikiConfig struct {
	BaseUrl  string
	UserName string
	Token    string
}

type SecretConfig struct {
	BaseUrl  string
	UserName string
	Token    string
}

func (g *GitConfig) GetRemoteUrl(repository string) string {
	if g.Project != "" {
		repository = fmt.Sprintf("%s/%s", g.Project, repository)
	}
	switch g.Protocol {
	case GitSSH:
		return fmt.Sprintf("git@%s:%s.git", g.Host, repository)
	case GitHTTPS:
		return fmt.Sprintf("https://%s/%s.git", g.Host, repository)
	default:
		return fmt.Sprintf("https://%s/%s.git", g.Host, repository)
	}
}

func (g *GitConfig) GetRepositoryPath(repository string) string {
	if g.Project != "" {
		return fmt.Sprintf("%s/%s", g.Project, repository)
	}
	return repository
}

func (g *GitConfig) GetRepositoryUrl(repository string) string {
	if g.Project != "" {
		return fmt.Sprintf("https://%s/%s/%s", g.Host, g.Project, repository)
	}
	return fmt.Sprintf("https://%s/%s", g.Host, repository)
}

type SetupCiCdConfig struct {
	RootDestinationsPath        string
	TemplatesRepository         string
	TemplatesRepositoryBranch   string
	TemplatesDestinationDir     string
	GitOpsRepository            string
	GitOpsRepositoryBranch      string
	GitOpsDestinationDir        string
	GitOpsToolsRepository       string
	GitOpsToolsRepositoryBranch string
	GitOpsToolsDestinationDir   string
	DefaultImageName            string
	DefaultImageTag             string
	ExternalConfigMap           bool
	ConfigMapRepository         string
	ConfigMapRepositoryBranch   string
	ConfigMapDestinationDir     string
	ApplicationMainBranch       string
	ApplicationDestinationDir   string
}

type Config struct {
	LogLevel      logger.LogLevel
	Http          *httpConfig
	WebSocket     *wsConfig
	Redis         *redisConfig
	Cors          *corsConfig
	SetupCiCd     *SetupCiCdConfig
	GitService    GitService
	GitConfig     *GitConfig
	WikiService   WikiService
	WikiConfig    *WikiConfig
	SecretService SecretService
	SecretConfig  *SecretConfig
}

func New() *Config {
	return loadConfigFromEnv()
}

func loadConfigFromEnv() *Config {
	godotenv.Load(".env")
	return &Config{
		LogLevel: getEnumEnvWithDefault[logger.LogLevel]("LOGLEVEL", logger.Error, logger.LogLevelFromString),
		Http: &httpConfig{
			Path: getEnvWithDefault("HTTP_PATH", "api"),
			Port: getEnvWithDefault("HTTP_PORT", "8080"),
		},
		WebSocket: &wsConfig{
			ReadBufferSize:   getIntEnvWithDefault("WEBSOCKET_READBUFFERSIZE", 1024),
			WriteBufferSize:  getIntEnvWithDefault("WEBSOCKET_WRITEBUFFERSIZE", 1024),
			HandshakeTimeout: getDurationEnvWithDefault("WEBSOCKET_HANDSHAKETIMEOUT", 0),
		},
		Redis: &redisConfig{
			Addr:     getEnvWithDefault("REDIS_ADDR", "localhost:6379"),
			Password: getEnvWithDefault("REDIS_PASSWORD", ""),
			DB:       getIntEnvWithDefault("REDIS_DB", 0),
		},
		Cors: &corsConfig{
			AllowedOrigins:   []string{getEnvWithDefault("CORS_ALLOWEDORIGINS", "http://localhost:3000")},
			AllowedMethods:   []string{getEnvWithDefault("CORS_ALLOWEDMETHODS", "*")},
			AllowHeaders:     []string{getEnvWithDefault("CORS_ALLOWHEADERS", "*")},
			ExposeHeaders:    []string{getEnvWithDefault("CORS_EXPOSEHEADERS", "Content-Length")},
			AllowCredentials: os.Getenv("CORS_ALLOWCREDENTIALS") == "true",
			MaxAge:           getDurationEnvWithDefault("CORS_MAXAGE", 12*time.Hour),
		},
		SetupCiCd: &SetupCiCdConfig{
			RootDestinationsPath:        getEnvWithDefault("SETUPCICD_ROOTDESTINATIONSPATH", "/tmp/setup-ci-cd/{{process-id}}"),
			TemplatesRepository:         getEnvWithDefault("SETUPCICD_TEMPLATESREPOSITORY", "devportal-templates"),
			TemplatesRepositoryBranch:   getEnvWithDefault("SETUPCICD_TEMPLATESREPOSITORYBRANCH", "develop"),
			TemplatesDestinationDir:     getEnvWithDefault("SETUPCICD_TEMPLATESDESTINATIONDIR", "/tmp/setup-ci-cd/{{process-id}}/templates"),
			GitOpsRepository:            getEnvWithDefault("SETUPCICD_GITOPSREPOSITORY", "git-ops"),
			GitOpsRepositoryBranch:      getEnvWithDefault("SETUPCICD_GITOPSREPOSITORYBRANCH", "develop"),
			GitOpsDestinationDir:        getEnvWithDefault("SETUPCICD_GITOPSDESTINATIONDIR", "/tmp/setup-ci-cd/{{process-id}}/git-ops"),
			GitOpsToolsRepository:       getEnvWithDefault("SETUPCICD_GITOPSTOOLSREPOSITORY", "git-ops-tools"),
			GitOpsToolsRepositoryBranch: getEnvWithDefault("SETUPCICD_GITOPSTOOLSREPOSITORYBRANCH", "develop"),
			GitOpsToolsDestinationDir:   getEnvWithDefault("SETUPCICD_GITOPSTOOLSDESTINATIONDIR", "/tmp/setup-ci-cd/{{process-id}}/git-ops-tools"),
			DefaultImageName:            getEnvWithDefault("SETUPCICD_DEFAULTIMAGENAME", "melquiadesrodrigues/template-api"),
			DefaultImageTag:             getEnvWithDefault("SETUPCICD_DEFAULTIMAGETAG", "latest"),
			ExternalConfigMap:           os.Getenv("SETUPCICD_EXTERNALCONFIGMAP") == "true",
			ConfigMapRepository:         getEnvWithDefault("SETUPCICD_CONFIGMAPREPOSITORY", "config-maps"),
			ConfigMapRepositoryBranch:   getEnvWithDefault("SETUPCICD_CONFIGMAPREPOSITORYBRANCH", "develop"),
			ConfigMapDestinationDir:     getEnvWithDefault("SETUPCICD_CONFIGMAPDESTINATIONDIR", "/tmp/setup-ci-cd/{{process-id}}/config-maps"),
			ApplicationMainBranch:       getEnvWithDefault("SETUPCICD_APPLICATIONMAINBRANCH", "master"),
			ApplicationDestinationDir:   getEnvWithDefault("SETUPCICD_APPLICATIONDESTINATIONDIR", "/tmp/setup-ci-cd/{{process-id}}/application"),
		},
		GitService: getEnumEnvWithDefault[GitService]("GITSERVICE", GitBitbucket, GitServiceFromString),
		GitConfig: &GitConfig{
			Host:     getEnvWithDefault("GITCONFIG_HOST", ""),
			UserName: getEnvWithDefault("GITCONFIG_USERNAME", ""),
			Token:    getEnvWithDefault("GITCONFIG_TOKEN", ""),
			Project:  getEnvWithDefault("GITCONFIG_PROJECT", ""),
			Protocol: getEnumEnvWithDefault[GitProtocol]("GITCONFIG_PROTOCOL", GitSSH, GitProtocolFromString),
		},
		WikiService: getEnumEnvWithDefault[WikiService]("WIKISERVICE", WikiConfluence, WikiServiceFromString),
		WikiConfig: &WikiConfig{
			BaseUrl:  getEnvWithDefault("WIKICONFIG_BASEURL", ""),
			UserName: getEnvWithDefault("WIKICONFIG_USERNAME", ""),
			Token:    getEnvWithDefault("WIKICONFIG_TOKEN", ""),
		},
		SecretService: getEnumEnvWithDefault[SecretService]("SECRETSERVICE", SecretVault, SecretServiceFromString),
		SecretConfig: &SecretConfig{
			BaseUrl:  getEnvWithDefault("SECRETCONFIG_BASEURL", ""),
			UserName: getEnvWithDefault("SECRETCONFIG_USERNAME", ""),
			Token:    getEnvWithDefault("SECRETCONFIG_TOKEN", ""),
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnvWithDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getDurationEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnumEnvWithDefault[T any](key string, defaultValue T, loader func(value string) T) T {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return loader(valueStr)
}
