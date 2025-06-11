# fibonacci

Fibonacci is a web service that generates Fibonacci sequences.

## Setup

The project uses:
- [devbox](https://www.jetify.com/docs/devbox/installing_devbox/) for managing dependencies and task execution.
- [orbstack](https://orbstack.dev/download) for running Docker/Kubernetes locally with automatic port forwarding, domain name access and remote debugging.
- [skaffold](https://skaffold.dev/docs/install/) for developing on Kubernetes with hot reload and minimal image builds.
- [hyperdx](https://github.com/hyperdxio/hyperdx/blob/main/LOCAL.md) for testing open telemetry instrumentation during development.

As long as devbox is used, only the Orbstack (or an equivalent) has to be installed manually. Everything else is fully managed by
devbox, including the Go toolchain and protobuf tooling.

Shell containing all of the dependencies can be accessed with the following command:
```shell
devbox shell
```

Commands that require dependencies managed by devbox can be executed without entering the shell by prefixing them with the following:
```shell
devbox run [command]
```

Starting the project in hot-reload loop can be accomplished with the following command:
```shell
devbox run dev
```

While the hot reload loop is running, hyperdx UI can be accessed by navigating to http://k8s.orb.local:8080.

Dependencies managed by devbox can be made accessible to the IDE of choice by installing a direnv plugin.

## Using

Orbstack handles [port-forwarding](https://docs.orbstack.dev/architecture#network) out of the box, so the services can be reached locally by their domain name regardless of the service type.

### In Browser

It's enough to just navigate to the following url:
```shell
http://api.fibonacci.svc.cluster.local:8081/api/v1/generate?length=32
```

### In Terminal

The following command can be used to call the Fibonacci service's REST API:
```shell
curl "http://api.fibonacci.svc.cluster.local:8081/api/v1/generate?length=32"
```

P.S. the following commands require giving terminal emulator permissions to access local network devices or else they fail with no route to host.

The following command can used to call the Fibonacci service's gRPC API:
```shell
devbox run curl --data '{"length": 32}' http://api.fibonacci.svc.cluster.local:8080/api.v1.Fibonacci/GenerateSequence
```

The following command can be used to check Fibonacci service's health:
```shell
devbox run health http://api.fibonacci.svc.cluster.local:8080
```
