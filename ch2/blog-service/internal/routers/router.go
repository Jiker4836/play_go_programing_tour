package routers

import (
	_ "block-service/docs"
	"block-service/global"
	"block-service/internal/middleware"
	"block-service/internal/routers/api"
	v1 "block-service/internal/routers/api/v1"
	"block-service/pkg/limiter"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"time"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	if global.ServerSetting.RunMode == "debug" {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	} else {
		r.Use(middleware.AccessLog())
		r.Use(middleware.Recovery())
	}
	methodLimiters := limiter.NewMethodLimiter().AddBuckets(
		limiter.LimitBucketRule{
			Key: "/auth",
			FillInterval: time.Second,
			Capacity: 10,
			Quantum: 10,
		})
	r.Use(middleware.RateLimiter(methodLimiters))
	r.Use(middleware.ContextTimeout(time.Second*60))
	r.Use(middleware.Translations())
	r.Use(middleware.JWT())

	r.GET("/swapper/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	upload := api.NewUpload()
	r.POST("/upload/file", upload.UploadFile)
	r.StaticFS("/static", http.Dir(global.AppSetting.UploadSavePath))
	r.GET("/auth", api.GetAuth)

	apiv1 := r.Group("/api/v1")

	article := v1.NewArticle()
	tag := v1.NewTag()
	apiv1.POST("/tags", tag.Create)
	apiv1.DELETE("/tags/:id", tag.Delete)
	apiv1.PUT("/tags/:id", tag.Update)
	apiv1.PATCH("/tags/:id/state", tag.Update)
	apiv1.GET("/tags", tag.List)

	apiv1.POST("/articles", article.Create)
	apiv1.DELETE("/articles/:id", article.Delete)
	apiv1.PUT("/articles/:id", article.Update)
	apiv1.PATCH("/articles/:id/state", article.Update)
	apiv1.GET("/articles/:id", article.Get)
	apiv1.GET("/articles", article.List)

	return r
}
