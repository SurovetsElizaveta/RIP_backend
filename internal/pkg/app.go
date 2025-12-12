package pkg

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"rip/internal/app/config"
	"rip/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

func (a *Application) RunHTTPS() {
	logrus.Info("HTTPS Server start up")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort),
		Handler: a.Router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	logrus.Infof("Starting HTTPS on https://%s:%d", a.Config.ServiceHost, a.Config.ServicePort)

	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("HTTPS Server down")
}
