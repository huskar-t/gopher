package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func CreateRouter(debug bool,corsConf *CorsConfig) *gin.Engine {
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	if debug {
		pprof.Register(router)
	}
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(cors.New(corsConf.GetConfig()))
	return router
}
