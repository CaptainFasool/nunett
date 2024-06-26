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
* `HeartbeatCollector` is a special collector for sending periodic heartbeats to a registered sink collector. A `HeartbeatCollector` can be anything, including another device-management-service or an `Allocation`.

See current reference model [collector.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/collector.go).

### Observable

The `Observable` interface determines how and which events happening in the system are to be observed by collectors. Any events that are eligible for observation at any `ObservabilityLevel` shall implement the interface. 

`ObservabilityLevel` is a constructed similarly to log levels concept in many programming languages and frameworks. We currently define six observability levels -- `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` and `FATAL`.

Any event that implements `Observable` interface can be observed by the collectors that are determined to be activated as a result of matching `ObservabilityLevel` declared by an event and the `ObservabilityLevel` environment variable.

See current reference model [observable.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/observable.go).


### Event

`Event` is an interface which defines methods to be implemented by a generic event of type `gEvent` and by that determines data that need to be included in each event for it be eligible to observation.

`EventCategory` is needed in order to account for the reasons for why we are doing observation of certain events and these are different from `ObservabilityLevel`s. Currently we are having the following event categories: `ACCOUNTING`, `LOGGING`, `TRACING`, `HEARTBEAT`. Note that there is clear relation between `EventCategory` and `Collector` types.

#### Generic Event `gEvent`

A generic event data type `gEvent` is then a type which joins together two interfaces -- `Event` and `Observable` -- and by that allows to a) gather all information needed to observe an event and b) direct the collection of that information to all registered collectors.

Developers are expected to choose to choose actions of the code to be considered events and observed at different levels by using `gEvent` implementations.

See current reference model [event.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/event.go).

### Message

_proposed by @kabir.kbr; 2024-04-16;_

`Message` is one of the key primitives of the NuNet platform -- the angle of the the architecture mostly influenced by the Actor model. This package defines `Message` interface and related data structures, but the transport-layer implementation of the interface is done in the [`network` package](../network/README.md). 

A `Message` interface defines two methods: `send()` and `receive()`. A generic type `gMessage` implements `Message` interface and requires necessary fields `sender`, `receiver`, `header` and `payload`.

See current reference model [message.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/telemetry/data/message.go).


### Heartbeat 

We define `Heartbeat` as a regularly send messages from the observed `Actor` entity in NuNet network (so `Node` or `Allocation`) to another entity, which can be another `Actor` or an external. This is done by registering a background process for all heartbeats of `Actor` associated with the same `Node` and triggering them per defined characteristics. A heartbeat is an `gEvent` of `HEARTBEAT` category which is being triggered periodically by a background process. It shall be implemented in `background_task` package which actually registering, maintain and stop heartbeat processes.

See the current reference model [heartbeat.go]()

#### Related functionalities

* [Request for heartbeat](#3-request-for-heartbeat) defined in this package;
* [Register heartbeat](../background_tasks/README.md#register-heartbeat) defined in `background_tasks` package;


## Functions

## 1. Register Collector(s)

_proposed 2024-04-08; by: @kabir.kbr;_

* signature: `dms.telemetry.registerCollector(gEvent gEvent, collector Collector) -> gEvent`;
* input #1: a variable of an generic event `dms.telemetry.gEvent` which will enable observation via registered `Collector`s.
* input #2: a variable describing collector to be registered `dms.telemetry.Collector` ([link](#collector));
* output type `dms.telemetry.gEvent` ([link](#event));

`registerCollector` function takes two parameters of type `gEvent` and `Collector` and outputs  `gEvent` type variable with the `Collector` registered in it.

* the main functionality of registering collectors is to automatically read default configuration of collectors and register them all into the generic event, which shall be used for instrumenting all events in the program -- see [Scenario: Register default collectors automatically](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/telemetry/registerCollector.feature);
* We also need to ensure that we could register collectors on demand, when they supplied in any manner (possibly also including via cli in the future) -- see: [Scenario: Register custom collectors manually](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/telemetry/registerCollector.feature)
   * The most important is the `OpenTelemetryCollector`, which will be used to collect data about all instrumented events;

## 2. Observe events

_proposed 2024-04-09; by: @kabir.kbr;_

The `gEvent` implements both `Event` and `Observable` interfaces which enables to mark each event in the program (chosen by a programmer) as an observable event and, provided the combination of `EventCategory` and `ObservabilityLevel` send telemetry information to registered collectors (e.g. `OpenTelemetryCollector`). 

* A correctly constructed event of `gEvent` type is observed by calling the `observeEvent()` method defined in `Observable` interface -- see [Feature: Observe gEvent](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/telemetry/observeEvent.feature).

## 3. Request for heartbeat

