# Introduction

The database package ontains configuration and functionality of database used by the DMS (Device Management Service).

# Stucture and organisation

Here is quick overview of the contents of this pacakge:

_Files_ 

* [README](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/README.md): Current file which is aimed towards developers who wish to use and modify the database functionality. 

* [generic_repository](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/generic_repository.go): This file defines the interface defining the main methods for database pacakge. It is designed using generic types and can be adapted to specific data type as needed.

* [generic_entity_repository](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/generic_entity_repository.go): This file contains the interface for those databases which will hold only a single record.

* [deployment](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/deployment.go): This file specifies a database interface having `DeploymentRequestFlat` data type.

* [elk_stats](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/elk_stats.go): This file specifies a database interface having `RequestTracker` data type.

* [errors](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/errors.go): This file specifies the different types of errors.

* [firecracker](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/firecracker.go): This file specifies a database interface having `VirtualMachine` data type.

* [machine](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/machine.go): This file defines database interfaces of various data types. 

* [onboarding](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/onboarding.go): This file specifies a database interface having `LogBinAuth` data type.

* [utils](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/utils.go): This file contains some utility functions with respect to database operations.

* [utils_test](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/utils_test.go): This file contains unit tests for functions defined in [utils.go](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/utils.go) file 

_Subpackages_

* [gorm](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/gorm): This folder contains SQlite database implementation using gorm.

* [clover](https://gitlab.com/nunet/device-management-service/-/blob/develop/db/repositories/clover): This folder contains CloverDB database implementation.

# Contributing

For guidelines of how to contribute, install and test the `device-management-service` component which contains `database` package, please refer to package level documentation:

* DMS component level [../README.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/README.md)
* Contribution guidelines [../CONTRIBUTING.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/CONTRIBUTING.md)
* Code of conduct [../CODE_OF_CONDUCT.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

# Specifications overview

* The specification of package functionality is described as test case definitions, maintained in repository [test-suite](https://gitlab.com/nunet/test-suite).
* The associated data models are specified and maintained in repository [open-api/platform-data-model/device-management-service/](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/develop/device-management-service/). 

Versioning and lifecycle of above mentioned specifications is aligned to the lifecycle and branching of the platform code (see [branching strategy](https://gitlab.com/nunet/documentation/-/wikis/GIT-Workflows#git-workflow-branching-strategy)):

* `develop` branches contain specifications of the functionality of current unstable branch of development at any given moment;
* `main` branches contain specifications of the current production version of the platform at any given moment in time;
* `proposed` branches contain new functionality and data model specifications, accepted for development, but not yet implemented.

The procedure to update the specifications is described in [Specification And Documentation Procedure](https://gitlab.com/nunet/team-processes-and-guidelines/-/blob/main/specification_and_documentation/README.md).

# Database Functionality

There are two types of interfaces defined to cover database operations:

1. `GenericRepository` 

2. `GenericEntityRepository`

These interfaces are described below.

## GenericRepository Interface

`GenericRepository` interface defines basic CRUD operations and standard querying methods. It is defined with generic data types. This allows it to be used for any data type. 

The methods of `GenericRepository` are as follows:

### Create

* signature: `Create(ctx context.Context, data T) -> (T, error)` <br/>

* input #1: Go context <br/>

* input #2: Data to be added to the database. It should be of type used to initialize the repository <br/>

* output (success): Data type used to initialize the repository <br/>

* output (error): error message

`Create` function adds a new record to the database. 

### Get

* signature: `Get(ctx context.Context, id interface{}) -> (T, error)` <br/>

* input #1: Go context <br/>

* input #2: Identifier of the record. Can be any data type <br/>

* output (success): Data with the identifier provided. It is of type used to initialize the repository <br/>

* output (error): error message

`Get` function retrieves a record from the database by its identifier. 

### Update

* signature: `Update(ctx context.Context, id interface{}, data T) -> (T, error)` <br/>

* input #1: Go context <br/>

* input #2: Identifier of the record. Can be any data type <br/>

* input #3: New data of type used to initialize the repository  <br/>

* output (success): Updated record of type used to initialize the repository <br/>

* output (error): error message

`Update` function modifies an existing record in the database using its identifier.

### Delete

* signature: `Delete(ctx context.Context, id interface{}) -> error` <br/>

* input #1: Go context <br/>

* input #2: Identifier of the record. Can be any data type <br/>

* output (success): None <br/>

* output (error): error message

`Delete` function deletes an existing record in the database using its identifier.

### Find

* signature: `Find(ctx context.Context, query Query[T]) -> (T, error)` <br/>

* input #1: Go context <br/>

* input #2: Query of type `dms.database.query` <br/>

* output (success): Result of query having the data type used to initialize the repository <br/>

* output (error): error message

`Find` function retrieves a single record from the database based on a query.

### FindAll

* signature: `FindAll(ctx context.Context, query Query[T]) -> ([]T, error)` <br/>

* input #1: Go context <br/>

* input #2: Query of type `dms.database.query` <br/>

* output (success): Lists of records based on query result. The data type of each record will be what was used to initialize the repository <br/>

* output (error): error message

`FindAll` function retrieves multiple records from the database based on a query.

### GetQuery
* signature: `GetQuery() -> Query[T]` <br/>

* input: None <br/>

* output: Query of type `dms.database.query`<br/>

`GetQuery` function returns an empty query instance for the repository's type.

## GenericEntityRepository Interface

`GenericEntityRepository` defines basic CRUD operations for repositories handling a single record. It is defined with generic data types. This allows it to be used for any data type. 

The methods of `GenericEntityRepository` are as follows:

### Save

* signature: `Save(ctx context.Context, data T) -> (T, error)` <br/>

* input #1: Go context <br/>

* input #2: Data to be saved of type used to initialize the database <br/>

* output (success): Updated record of type used to initialize the repository <br/>

* output (error): error message

`Save` function adds or updates a single record in the repository

### Get
* signature: `Get(ctx context.Context) -> (T, error)` <br/>

* input: Go context <br/>

* output (success): Record of type used to initialize the repository <br/>

* output (error): error message

`Get` function retrieves the single record from the database.

### Clear

* signature: `Clear(ctx context.Context) -> error` <br/>

* input: Go context <br/>

* output (success): None <br/>

* output (error): error 

`Clear` function removes the record and its history from the repository.

### History

* signature: `History(ctx context.Context, qiery Query[T]) -> ([]T, error)` <br/>

* input #1: Go context <br/>

* input #2: query of type `dms.database.query` <br/>

* output (success):List of records of repository's type <br/>

* output (error): error 

`History` function retrieves previous records from the repository which satisfy the query conditions.

### GetQuery

* signature: `GetQuery() -> Query[T]` <br/>

* input: None <br/>

* output: New query of type `dms.database.query` <br/>

`GetQuery` function returns an empty query instance for the repository's type.

## List of Data Types

`dms.database.query`: This contains parameters related to a query that is passed to the database. See [query.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/database-spec/device-management-service/database/data/query.data.go) for reference data model.
