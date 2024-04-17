# Introduction
This sub-package contains functionality including drivers and api for the Docker executor.

# Stucture and organisation

Here is quick overview of the contents of this pacakge:

* [README](README.md): File which is aimed towards developers who wish to use and modify the docker functionality. 

* [client](client.go): 

* [executor](executor.go): This file contains 

* [handler](handler.go): This file contains 

* [init](init.go): This file contains endpoints 

* [types](types.go): This file contains  

# Contributing

For guidelines of how to contribute, install and test the `device-management-service` component which contains `executor` package, please refer to package level documentation:

* Executor package level [../README.md](../README.md)
* DMS component level [../../README.md](../../README.md)
* Contribution guidelines [../../CONTRIBUTING.md](../../CONTRIBUTING.md)
* Code of conduct [../../CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

# Specifications overview

* The specification of package functionality is described as test case definitions, maintained in repository [test-suite](https://gitlab.com/nunet/test-suite).
* The associated data models are specified and maintained in repository [open-api/platform-data-model/device-management-service/](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/develop/device-management-service/). 

Versioning and lifecycle of above mentioned specifications is aligned to the lifecycle and branching of the platform code (see [branching strategy](https://gitlab.com/nunet/documentation/-/wikis/GIT-Workflows#git-workflow-branching-strategy)):

* `develop` branches contain specifications of the functionality of current unstable branch of development at any given moment;
* `main` branches contain specifications of the current production version of the platform at any given moment in time;
* `proposed` branches contain new functionality and data model specifications, accepted for development, but not yet implemented.

The procedure to update the specifications is described in [Specification And Documentation Procedure](https://gitlab.com/nunet/team-processes-and-guidelines/-/blob/main/Specification_And_Documentation_Procedure.md?ref_type=heads).


