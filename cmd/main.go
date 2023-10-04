package main

import (
	"context"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	redisPkg "github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/userpass"
	bitbucketPkg "github.com/ktrysmt/go-bitbucket"
	"github.com/zahirsis/dev-portal-backend/config"
	confluenceapi "github.com/zahirsis/dev-portal-backend/pkg/confluence-api-v2"
	"github.com/zahirsis/dev-portal-backend/pkg/log_logger"
	httpHandler "github.com/zahirsis/dev-portal-backend/src/app/handlers/http"
	websocketHandler "github.com/zahirsis/dev-portal-backend/src/app/handlers/websocket"
	"github.com/zahirsis/dev-portal-backend/src/app/usecase"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/repository/memory"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/repository/redis"
	awsApp "github.com/zahirsis/dev-portal-backend/src/infrastructure/services/aws"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/services/bitbucket"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/services/confluence"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/services/unix"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/services/vault"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"github.com/zahirsis/dev-portal-backend/src/pkg/messenger"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx := context.TODO()
	cfg := config.New()
	loggerInstance := log_logger.New(log.New(os.Stdout, "", log.Ldate|log.Ltime), &logger.Config{Level: cfg.LogLevel})
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Cors.AllowedOrigins,
		AllowMethods:     cfg.Cors.AllowedMethods,
		AllowHeaders:     cfg.Cors.AllowHeaders,
		ExposeHeaders:    cfg.Cors.ExposeHeaders,
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))
	apiGroup := router.Group(cfg.Http.Path)
	redisClient := redisPkg.NewClient(&redisPkg.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	upgrader := &websocket.Upgrader{
		ReadBufferSize:   cfg.WebSocket.ReadBufferSize,
		WriteBufferSize:  cfg.WebSocket.WriteBufferSize,
		HandshakeTimeout: cfg.WebSocket.HandshakeTimeout,
		WriteBufferPool:  nil,
		Subprotocols:     nil,
		Error:            nil,
		CheckOrigin: func(request *http.Request) bool {
			for _, origin := range cfg.Cors.AllowedOrigins {
				loggerInstance.Debug("Checking origin", origin, request.Header.Get("Origin"))
				if request.Header.Get("Origin") == origin {
					return true
				}
			}
			return false
		},
		EnableCompression: false,
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		loggerInstance.Fatal("Error loading AWS config", err)
		return
	}
	awsClient := ecr.NewFromConfig(awsCfg)
	messengerInstance := messenger.NewMessageMassager()

	tr := memory.NewTemplateRepository(loggerInstance)
	er := memory.NewEnvironmentRepository(loggerInstance)
	sr := memory.NewSquadRepository(loggerInstance)
	pr := redis.NewProcessRepository(loggerInstance, redisClient)
	mr := memory.NewManifestRepository(loggerInstance)
	rc := &repository.Container{
		TemplateRepository:    tr,
		EnvironmentRepository: er,
		SquadRepository:       sr,
		ProgressRepository:    pr,
		ManifestRepository:    mr,
	}

	bitbucketClient := bitbucketPkg.NewBasicAuth(cfg.GitConfig.UserName, cfg.GitConfig.Token)
	confluenceApi, err := confluenceapi.NewAPI(cfg.WikiConfig.BaseUrl, cfg.WikiConfig.UserName, cfg.WikiConfig.Token)
	if err != nil {
		loggerInstance.Fatal("Error creating confluence API", err)
		return
	}
	vdc := api.DefaultConfig()
	vdc.Address = cfg.SecretConfig.BaseUrl
	vaultApi, err := api.NewClient(vdc)
	if err != nil {
		loggerInstance.Fatal("Error creating vault API", err)
		return
	}
	vaultAuth, err := userpass.NewUserpassAuth(cfg.SecretConfig.UserName, &userpass.Password{
		FromString: cfg.SecretConfig.Token,
	})
	if err != nil {
		loggerInstance.Fatal("Error creating vault auth", err)
		return
	}

	git := unix.NewGitService(cfg.GitConfig, loggerInstance)
	ccs := service.NewCiCdService(loggerInstance, rc)
	rs := service.NewRegistryService(loggerInstance, rc)
	ras := awsApp.NewRegistryApiService(loggerInstance, awsClient)
	ds := unix.NewDirectoryService(loggerInstance)
	gs := service.NewGitOpsService(cfg, loggerInstance, ds)
	ps := service.NewPipelineService(cfg, loggerInstance, ds)
	gas := bitbucket.NewGitApiService(cfg.GitConfig, loggerInstance, bitbucketClient)
	aws := confluence.NewConfluenceService(cfg, loggerInstance, confluenceApi)
	ws := service.NewWikiService(cfg, loggerInstance, aws, ds)
	sas := vault.NewSecretApiService(cfg, loggerInstance, vaultApi, vaultAuth)
	ss := service.NewSecretService(loggerInstance, sas)
	sc := &service.Container{
		GitService:         git,
		CiCdService:        ccs,
		RegistryService:    rs,
		RegistryApiService: ras,
		GitOpsService:      gs,
		PipelineService:    ps,
		DirectoryService:   ds,
		GitApiService:      gas,
		WikiService:        ws,
		WikiApiService:     aws,
		SecretService:      ss,
		SecretApiService:   sas,
	}
	c := &container.Container{
		Logger:         loggerInstance,
		MessageManager: messengerInstance,
		Repositories:   rc,
		Services:       sc,
	}

	// Templates
	tuc := usecase.NewListTemplatesUseCase(c)
	httpHandler.NewTemplateHandler(c, apiGroup.Group("templates"), tuc)

	// Environments
	euc := usecase.NewListEnvironmentsUseCase(c)
	httpHandler.NewEnvironmentHandler(c, apiGroup.Group("environments"), euc)

	// Squads
	suc := usecase.NewListSquadsUseCase(c)
	httpHandler.NewSquadHandler(c, apiGroup.Group("squads"), suc)

	// CI/CD
	cuc := usecase.NewSetupCiCdUseCase(c, cfg)
	httpHandler.NewCiCdHandler(c, apiGroup.Group("ci-cd"), cuc)
	puc := usecase.NewProgressUseCase(c)
	websocketHandler.NewCiCdHandler(c, apiGroup.Group("ci-cd"), upgrader, puc)

	err = router.Run(":8080")
	if err != nil {
		loggerInstance.Error("Error running server", err)
		return
	}
}
