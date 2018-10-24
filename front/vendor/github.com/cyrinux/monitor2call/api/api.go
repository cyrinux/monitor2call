package api

// Les imports de librairies
import (
	"github.com/cyrinux/monitor2call/helpers" // common helpers
	_ "github.com/cyrinux/monitor2call/front/docs" //swagger documentation
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"

	"github.com/swaggo/gin-swagger" //swagger contollers
	"github.com/swaggo/gin-swagger/swaggerFiles" //swag files
)


// Handlers contains routes
func Handlers() *gin.Engine {
	ginMode := helpers.GetEnv("GIN_MODE", "debug")
	readPassword := helpers.GetEnv("READ_PASSWORD", "")
	writePassword := helpers.GetEnv("WRITE_PASSWORD", "")
	cacheDir := helpers.GetEnv("CACHE_DIR", "./cache")

	// init logrus logger
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	// Creates a router without any middleware by default
	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery(), Cors())

	// enable AUTH on RELEASE mode
	if ginMode != "release" {

		// serve files
		r.Static("/statics", cacheDir)
		r.Static("/assets", "./assets")

		// user endpoint
		v1Users := r.Group("api/v1/users")
		{
			v1Users.POST("", PostUser)
			v1Users.GET("", GetUsers)
			v1Users.GET(":id", GetUser)
			// v1Users.PUT(":id", EditUser)
			v1Users.DELETE(":id", DeleteUser)
			v1Users.OPTIONS("", OptionsUser)      // POST
			v1Users.OPTIONS(":id", OptionsUser) // PUT, DELETE
		}

		// alert endpoint
		v1Alerts := r.Group("api/v1/alerts")
		{
			v1Alerts.POST("", PostAlert)
			v1Alerts.GET("", GetAlerts)
			v1Alerts.GET(":id", GetAlert)
			v1Alerts.PUT(":id", PostAlert)
			v1Alerts.DELETE(":id", DeleteAlert)
			v1Alerts.OPTIONS("", OptionsAlert)    // POST
			v1Alerts.OPTIONS(":id", OptionsAlert) // PUT, DELETE
		}
	} else {

		// serve files
		v1Assets := r.Group("/", gin.BasicAuth(gin.Accounts{
			"read":  readPassword,
			"write": writePassword,
		}))
		{
			v1Assets.Static("/statics", cacheDir)
			v1Assets.Static("/assets", "./assets")
		}
		

		// user read endpoint
		v1UsersRead := r.Group("api/v1/users", gin.BasicAuth(gin.Accounts{
			"read": readPassword,
		}))
		{
			v1UsersRead.GET("", GetUsers)
			v1UsersRead.GET(":id", GetUser)
		}

		// user write endpoint
		v1UsersWrite := r.Group("api/v1/users", gin.BasicAuth(gin.Accounts{
			"write": writePassword,
		}))
		{
			v1UsersWrite.POST("", PostUser)
			// v1UsersWrite.PUT(":id", EditUser)
			v1UsersWrite.DELETE(":id", DeleteUser)
			v1UsersWrite.OPTIONS("", OptionsUser)      // POST
			v1UsersWrite.OPTIONS(":id", OptionsUser) // PUT, DELETE
		}

		// read alert endpoint
		v1AlertsRead := r.Group("api/v1/alerts", gin.BasicAuth(gin.Accounts{
			"read": readPassword,
		}))
		{
			v1AlertsRead.GET("", GetAlerts)
			v1AlertsRead.GET(":id", GetAlert)
		}

		// write alert endpoint
		v1AlertsWrite := r.Group("api/v1/alerts", gin.BasicAuth(gin.Accounts{
			"write": writePassword,
		}))
		{
			v1AlertsWrite.POST("", PostAlert)
			v1AlertsWrite.PUT(":id", PostAlert)
			v1AlertsWrite.DELETE(":id", DeleteAlert)
			v1AlertsWrite.OPTIONS("", OptionsAlert)    // POST
			v1AlertsWrite.OPTIONS(":id", OptionsAlert) // PUT, DELETE
		}
	}

	// swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// Cors enabled
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}
