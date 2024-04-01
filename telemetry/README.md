# Introduction

This package contains logs, tracers and everything related to telemetry. The design and implementation of this package has been initiated within milesotne []In order to implement all functions below, we first define a few datatypes and interfaces.

## Interfaces and types

### Collector

A collector is a data sink that collects data about observed events. Which events it is ovserving and what does it do to process data that it collects is entirely upon the logic of the collector. Here we just define the interface for any collector to be able to registered on any `dms`.

The goal is to allow third parties to propose their collectors and for `dms` to be able to register them and observe required events via implementation of `Observable` interface. For now, we consider three types of `Collector`s: 

* `FileCollector`, collecting events into local file;
* `DatabaseCollector`, collecting events into local database;
* `OpenTelemetryCollector`, sending telemetry events to the registered open telemetry collector using open-telemetry format;
* `ReputationCollector` is not scheduled for the implementation now, but is considered for the future to be able to register reputation systems that will collect events for providing reputation services on the platform; 

### Observable

### Event

### Message

## LocalEvent

## Functions

## 1. Register Collector

_proposed 2024-04-01; by: @kabir.kbr;_



Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/orchestrator/Job_Posting.feature))   |
| Request payload       | [jobDescription](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/jobs/data/jobDescription.payload.go)|
| Return payload       | None |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/jobPosting.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/rendered/jobPosting.sequence.svg)) | 






