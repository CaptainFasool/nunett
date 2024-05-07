# Introduction

The integrations package should contain third party integrations.



** TODO: all content in this package and README content below needs to be moved to its appropriate package 
    integrations package should contain code that's more aligned with the conventional integrations
    packages - third party software integration.
**


**How to compile proto file?**

1. Install `protoc`. See [Protocol Buffer Compiler Installation](https://grpc.io/docs/protoc-installation/).

2. Install Go specific plugins for protoc.

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

3. Run following command from the directory this README file is in:

```
protoc --go_out=.  --go-grpc_out=.  specs/compute-api-spec/oracle.proto
```

4. The contents/symbols of the generated content can be referenced from `gitlab.com/nunet/device-management-service/integrations/oracle`.

