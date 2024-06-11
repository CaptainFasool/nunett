# Introduction

This package contains all configuration related code such as reading config file and functions to configure at runtime.

There are two sides to configuration:
1. default configuration, which has to be loaded for the fresh installation of new dms;
2. dynamic configuration, which can be changed by a user that has access to the DMS; this dynamic configuration may need to be persistent (or not).

## Default configuration

Default configuration should be included into DMS distribution as a `config.yaml` in the root directory. The following is loosely based on general practice of passing yaml configuration to Go programs (see e.g. [A clean way to pass configs in a Go application](https://dev.to/ilyakaznacheev/a-clean-way-to-pass-configs-in-a-go-application-1g64)). DMS would parse this file during onboarding and populate the `internal.config.Config` variable that will be imported to other packages and used accordingly. 

## Dynamic configuration

Dynamic configuration would use the same `internal.config.Config` variable, but would allow for adding new values or changing configuration by an authorized DMS user -- via DMS CLI or REST API calls. 

The mechanism of dynamic configuration will enable to override or change default values. For enabling this functionality, the `internal.config.Config` variable will have a synchronized copy in the local DMS database, defined with `dms.database` package. 

## Functionality

1. Load default DMS configuration; see [scenario definition](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/internal/config/configurationManagement.feature?ref_type=heads#L3);
2. Restore default DMS configuration; see [scenario definition](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/internal/config/configurationManagement.feature?ref_type=heads#L14);
3. Load existing DMS configuration; see [scenario definition](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/internal/config/configurationManagement.feature?ref_type=heads#L28);