package config

import (
	"os"

	"github.com/SokolovAleksandr/AvitoTestBackend/internal/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "data/config.yaml"

type DatabaseParams struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func (p *DatabaseParams) GetHost() string {
	return p.Host
}

func (p *DatabaseParams) GetPort() int {
	return p.Port
}

func (p *DatabaseParams) GetUser() string {
	return p.User
}

func (p *DatabaseParams) GetPassword() string {
	return p.Password
}

func (p *DatabaseParams) GetDBName() string {
	return p.DBName
}

type Config struct {
	HttpPort int             `yaml:"port"`
	DB       *DatabaseParams `yaml:"database"`
}

type Service struct {
	config Config
}

func New() (*Service, error) {
	s := &Service{}

	logger.Debug("reading config...")
	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}
	logger.Debug("reading config finished")

	logger.Debug("unmarshaling config...")
	err = yaml.Unmarshal(rawYAML, &s.config)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}
	logger.Debug("unmarshaling config finished")

	return s, nil
}

func (s *Service) GetRepositoryParams() (*DatabaseParams, error) {
	return s.config.DB, nil
}

func (s *Service) GetHttpPort() (int, error) {
	return s.config.HttpPort, nil
}
