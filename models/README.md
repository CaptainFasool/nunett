# Introduction 

Models package defines and keeps data structures and interfaces that are used across the whole dms component by different packages. 

# Models

## Actor

_proposed by: @kabir.kbr; date: 2024-04-17_

An Actor is an entity in NuNet ontology and nomenclature (see [research blog]](https://nunet.gitlab.io/research/blog/posts/ontology-and-nomenclature/#actors)), which: 
1. can be uniquely identified in the network;
2. can receive messages from other actors;
3. can send messages to other actors;
4. has behaviors that can be triggered by received messages (so implements a remote procedure calls between Actors).

Actor interface is implemented by two actors which functionality is determined by device-management-service package: `Node` (in `dms` package,) and `Allocation` (in `jobs` package). 

See proposed `Actor` interface specification in [nunet/open-api/platform-data-model/models/actor.go](https://gitlab.com/nunet/open-api/platform-data-model/models/actor.go)

## ID

Defines data types that are used to uniquely identify Entities in the network. 