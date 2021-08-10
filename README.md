# Condo CLI
CLI for the condo build system github.com/automotiveMastermind/condo

## Install instructions: 

build the project
```
go build
```

link the built binary to the path (OSX)
```
ln -sf $(pwd)/condo-cli /usr/local/bin/condo
```

then run!
```
condo run
```


## Cluster dependencies: 
kind
git
kubectl
helm
docker
