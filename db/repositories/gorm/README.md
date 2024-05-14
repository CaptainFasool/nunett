# Introduction

This sub package contains Gorm implementation of the database interfaces.

# Stucture and organisation

Here is quick overview of the contents of this pacakge:

* [README](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/README.md): Current file which is aimed towards developers who wish to use and modify the database functionality. 

* [generic_repository](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/generic_repository.go): This file implements the methods of `GenericRepository` interface.

* [generic_entity_repository](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/generic_entity_repository.go): This file implements the methods of `GenericEntityRepository` interface.

* [deployment](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/deployment.go): This file contains implementation of `DeploymentRequestFlatRepository` interface. 

* [elk_stats](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/elk_stats.go): This file contains implementation of `RequestTrackerRepository` interface.

* [firecracker](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/firecracker.go): This file contains implementation of `VirtualMachineRepository` interface.

* [machine](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/machine.go): This file contains implementation of interfaces defined in [machine.go](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/machine.go).  

* [onboarding](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/onboarding.go): This file contains implementation of `LogBinAuthRepository` interface.

* [utils](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm/utils.go): This file contains utility functions with respect to Gorm implementation.

All files with `*_test.go` naming convention contain unit tests with respect to the specific implementation.

# Contributing

For guidelines of how to contribute, install and test the `device-management-service` component which contains `database` package, please refer to package level documentation:

* Database package level [../README.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md)
* DMS component level [../../README.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/README.md)
* Contribution guidelines [../../CONTRIBUTING.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/CONTRIBUTING.md)
* Code of conduct [../../CODE_OF_CONDUCT.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

# Specifications overview

* The specification of package functionality is described as test case definitions, maintained in repository [test-suite](https://gitlab.com/nunet/test-suite).
* The associated data models are specified and maintained in repository [open-api/platform-data-model/device-management-service/](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/develop/device-management-service/). 

Versioning and lifecycle of above mentioned specifications is aligned to the lifecycle and branching of the platform code (see [branching strategy](https://gitlab.com/nunet/documentation/-/wikis/GIT-Workflows#git-workflow-branching-strategy)):

* `develop` branches contain specifications of the functionality of current unstable branch of development at any given moment;
* `main` branches contain specifications of the current production version of the platform at any given moment in time;
* `proposed` branches contain new functionality and data model specifications, accepted for development, but not yet implemented.

The procedure to update the specifications is described in [Specification And Documentation Procedure](https://gitlab.com/nunet/team-processes-and-guidelines/-/blob/main/specification_and_documentation/README.md).

# Functions

## GenericRepository

### NewGenericRepository

* signature: `NewGenericRepository[T repositories.ModelType](db *gorm.DB) -> repositories.GenericRepository[T]` <br/>

* input: Gorm Database object <br/>

* output: Repository of type `dms.database.gorm.GenericRepositoryGORM` <br/>

`NewGenericRepository` function creates a new instance of `GenericRepositoryGORM` struct. It initializes and returns a repository with the provided GORM database. 

### Create

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#create) 

`Create` function adds a new record to the database and returns the created data. It returns an error message in case of any error during the operation.

See [Feature: Creating a record in the repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Create.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/create.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/create.sequence.svg)

### Get

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#get) 

`Get` function retrieves a record from the database based on the identifier provided. It returns an error message in case of any error during the operation.

See [Feature: Retrieving a record from the repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Get.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/get.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/get.sequence.svg)

### Update

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#update) 

`Update` function modifies an existing record in the database using its identifier. It returns an error message in case of any error during the operation.

See [Feature: Updating a record in the repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Update.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/update.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/update.sequence.svg)

### Delete

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#delete) 

`Delete` function deletes an existing record in the database using its identifier. It returns an error message in case of any error during the operation.

See [Feature: Deleting a record from the repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Delete.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/delete.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/delete.sequence.svg)

### Find

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#find) 

`Find` function retrieves a single record from the database based on a query. It returns an error message in case of any error during the operation.

See [Feature: Retrieving a single record based on a query from the repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Find.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/find.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/find.sequence.svg)

### FindAll

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#findall) 

`FindAll` function retrieves multiple records from the database based on a query. It returns an error message in case of any error during the operation.

See [Feature: Retrieving multiple records based on a query from the repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Find_All.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/findAll.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/findAll.sequence.svg)

### GetQuery

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#getquery) 

`GetQuery` function returns a clean query instance for the repository's type.

See [Feature: Getting a clean Query instance of repository's type](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Get_Query.feature) for test cases.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/getQuery.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/getQuery.sequence.svg)

## GenericEntityRepository

### NewGenericEntityRepository

* signature: `NewGenericEntityRepository[T repositories.ModelType](db *gorm.DB) -> repositories.GenericEntityRepository[T]` 

* input #1: Gorm Database object <br/>

* output: Repository of type `dms.database.gorm.GenericEntityRepositoryGORM` <br/>

`NewGenericEntityRepository` creates a new instance of `GenericEntityRepositoryGORM` struct. It initializes and returns a repository with the provided GORM database.

### Save

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#save) 

`Save` creates or updates the record in the repository and returns the new/updated data. It returns an error message in case of any error during the operation.

See [Feature: Saving a record in the entity repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Save_Entity.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/saveEntity.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/saveEntity.sequence.svg)

### Get

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#get-1) 

`Get` function retrieves the single record from the database. It returns an error message in case of any error during the operation.

See [Feature: Retrieving a record from the entity repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Get_Entity.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/getEntity.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/getEntity.sequence.svg)

### Clear

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#clear) 

`Clear` function removes the record with its history from the repository. It returns an error in case of any error during the operation.

See [Feature: Removing record from the entity repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Clear_Entity.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/clearEntity.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/clearEntity.sequence.svg)

### History

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md#history) 

`History` function retrieves previous versions of the record from the repository which meet the provided query conditions. It returns an error in case of any error during the operation.

See [Feature: Retrieving History from Entity Repository](https://gitlab.com/nunet/test-suite/-/blob/database-spec/stages/functional_tests/features/device-management-service/database/gorm/Clear_Entity.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/historyEntity.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/sequences/rendered/historyEntity.sequence.svg)

## List of Data Types

`dms.database.genericRepositoryGORM`: This is a generic repository implementation using GORM as an ORM. See [genericRepositoryGORM.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/data/genericRepositoryGORM.data.go) for reference data model.

`dms.database.genericEntityRepositoryGORM`: This is a generic single entity repository implementation using GORM as an ORM. See [genericEntityRepositoryGORM.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/gorm/data/genericEntityRepositoryGORM.data.go) for reference data model.
