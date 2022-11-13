package main

import (
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/config"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/logger"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/server"
)

func main() {
	logger.Info("initing backend...")

	config, err := config.New()
	if err != nil {
		logger.Fatal("config init failed", "error", err.Error())
	}

	port, err := config.GetHttpPort()
	if err != nil {
		logger.Fatal("extract port failed", "error", err.Error())
	}

	repParams, err := config.GetRepositoryParams()
	if err != nil {
		logger.Fatal("extract repository params failed", "error", err.Error())
	}

	app, err := server.New(port, repParams)
	if err != nil {
		logger.Fatal("server init failed", "error", err.Error())
	}

	logger.Info("initing back finished.")

	logger.Info("running back...")
	if err := app.Run(); err != nil {
		logger.Fatal("running back failed", "error", err.Error())
	}
	logger.Info("running back finished")
}
