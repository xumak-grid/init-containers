# Nexus init configuration

This is a k8s job to create initial configuration for Nexus 3.x server.
The grid/init-nexus container reads the configuration file to configure Nexus making POST requests to its API

## Secret

The secret contains the configuration that will be used to the job, it is a json config that contains slices of Group, Hosted and Proxy repositories (see examples/configFile.json)

You can create the secret using the example file

```
kubectl create secret generic nexus-init-config --from-file=./examples/configFile.json

```

or using the yaml file, the content of the file is passed in base64 encode to obtain the value


```
cat examples/configFile.json | base64
kubectl apply -f k8s/secret.yaml
```

## Job

The job reads the configFile.json and start posting requests to the host (Nexus) see the Makefile for more info

`kubectl apply -f k8s/job.yaml`

### Local test

The init container contains default values for the following env vars.

```
NEXUS_USER="admin"
NEXUS_PASS="admin123"
NEXUS_HOST="http://localhost:8081"
// this file location contains configuration to make a initial setup to Nexus server
NEXUS_CONFIG_FILE="examples/configFile.json"
```

```
docker run \
    --rm -d --name nexus -p 8081:8081 \
    281327226678.dkr.ecr.us-east-1.amazonaws.com/grid/nexus:3.12.0
```

`go run main.go config.go`
