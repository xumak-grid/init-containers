# Gogs init configuration

This is a k8s job to create initial configuration for Gogs 0.11.x server.
The grid/init-gogs container reads the configuration file to configure Gogs making POST requests to its API

## Secret

The secret contains the configuration that will be used to the job, it is a json config that contains slice of init_data, organizations and repostitories (see examples/configFile.json)

You can create the secret using the example file

```
kubectl create secret generic gogs-init-config --from-file=./examples/configFile.json

```

or using the yaml file, the content of the file is passed in base64 encode to obtain the value


```
cat examples/configFile.json | base64
kubectl apply -f k8s/secret.yaml
```

## Job

The job reads the configFile.json and start posting requests to the host (Gogs) see the Makefile for more info

`kubectl apply -f k8s/job.yaml`

Inside k8s the following values in configFile.json are important e.g.

```
    "domain": "gogs-server",
    "app_url": "",
```

## Clone the demo repository

For demo purpose a new deploy key is added inside the container this allows to clone xumak-grid/demo
`ssh-keygen -t rsa -N "" -f $PWD/id_rsa -C ""`


## Local test

This repository provides a [docker-compose](docker-compose.yml) file in order to allow create a local test environment.

### Start gogs service

    docker-compose up -d gogs

### Start init configuration service

    docker-compose up init

Go to the following URL to see the Gogs environment:

    http://localhost:8181/

### Environment vars
The docker-compose file contains the following environment vars:
* GOGS_HOST: It should be the url where the gogs server is exposed.
* GOGS_CONFIG_FILE: File path that contains the gogs configuration. The init-gogs image already contains some configuration files in order to test.

It is not necessary to change any default value in order to make a local test.