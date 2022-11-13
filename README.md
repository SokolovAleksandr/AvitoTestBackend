# AvitoTestBackend

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)

## Description

This project is a solution to Avito Backend Test Task.

Description of Avito Backend Test Task is stored in [README_task.me](README_task.md).

Application uses PG database.

## Configuration

Application uses config file on path: data/config.yaml.

Application config example is stored in [config.example.yaml](config.example.yaml).

## Build Application

To build application execute:

```bash
make build
```

````

## Run Application

To run application locally execute:

```bash
make run
```

# Run Application in Docker

To run application in Docker execute

```bash
make docker
```

Make sure that Docker file exposes same port as presented in config file.

## Usage

Appliication usage examples are stored as [PostMan file](AvitoTestBackend_Requests.postman_collection.json).

Swagger docs for API are presented on [/swagger](http://localhost:8080/swagger).
Also swagger docs can be viewed as json in [swagger.json](docs/swagger.json).
````
