# Introduction

This package contains logs, tracers and everything related to telemetry. The design and implementation of this package has been initiated within milesotne []In order to implement all functions below, we first define a few datatypes and interfaces.

## Interfaces and types

* _proposed 2024-04-02; by: @kabir.kbr;_

### Collector

A collector is a data sink that collects data about observed events. Which events it is ovserving and what does it do to process data that it collects is entirely upon the logic of the collector. Here we just define the interface for any collector to be able to registered on any `dms`.

The goal is to allow third parties to propose their collectors and for `dms` to be able to register them and observe required events via implementation of `Observable` interface. For now, we consider three types of `Collector`s: 

* `FileCollector`, collecting events into local file;
* `DatabaseCollector`, collecting events into local database;
* `OpenTelemetryCollector`, sending telemetry events to the registered open telemetry collector using open-telemetry format;
* `ReputationCollector` is not scheduled for the implementation now, but is considered for the future to be able to register reputation systems that will collect events for providing reputation services on the platform; 

See current reference model [collector.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/collector.go).

### Observable

The `Observable` interface determines how and which events hapenning in the system are to be observed by collectors. Any events that are eligible for observation at any `ObservabilityLevel` shall implement the interface. 

`ObservabilityLevel` is a constructed similarly to log levels concept in many programming languages and frameworks. We currently define six observability levels -- `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` and `FATAL`. 

Any event that implements `Observable` interface can be observed by the collectors that are determined to be activated as a result of matching `ObservabilityLevel` declared by an event and the `ObservabilityLevel` environment variable.

See current reference model [observable.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/observable.go).


### Event

`Event` is an interface which defines methods to be implemented by a generic event of type `gEvent` and by that determines data that need to be included in each event for it be eligible to observation.

A generic event data type `gEvent` is then a type which joins together two interfaces -- `Event` and `Observable` -- and by that allows to a) gahter all information needed to observe an event and 2) direct the collection of that information to all registered collectors.

Developers are expected to choose to choose actions of the code to be considered events and observed at different levels by using `gEvent` implementations.

See current reference model [event.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/event.go).


### Message

A `Message` is a special event of which defines a message between two `dms`es running in the network. In terms of Actor model, `Message` is one of the key primitives of the whole system. Besides being a special type of generic event `Message` extends `gEvent` by adding neccessary fields `sender`, `receiver`, `header` and `payload`, which may get some of their structures via `gEvent.context()` method.

All `Messages` should have a capability to be observed via `OpenTelemetryCollector` implementation -- extending `ObservabilityLevel` concept and type for determining when and how an observation takes place (e.g. during debugging, testing or in production). 

See current reference model [message.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/message.go).

## LocalEvent

`LocalEvent` type is an `gEvent` that is not a `Message` but still considered important enough to be observed. 

See current reference model [local_event.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/local_event.go).


## Functions

## 1. Register Collector

_proposed 2024-04-01; by: @kabir.kbr;_
_**TBD**_


Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin]())   |
| Request payload       | []()|
| Return payload       | None |
| Processes / Functions | sequenceDiagram ([.mermaid](),[.svg]()) | 






