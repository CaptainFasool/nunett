# Introduction

This package is for scheduling background tasks to preserve resources. Takes care of background tasks scheduling and execution, other packages that have their own background tasks register through this package:
1. Registration 
    1. The task itself, the arguments it needs
    2. priority 
    3. event (time period or other event to trigger task)
2. Start , Stop, Resume
3. Algorithm that accounts for the event and priority of the task (not yet clear) 
4. Monitor resource usage of tasks (not yet clear)

## Functions

### Register Heartbeat

_proposed by: @kabir.kbr; date: 2024-04-17_

TBD, required by `telemetry` package

See currently proposed interfaces and data model [heartbeat.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/background_tasks/heartbeat.go).

### Register `receiveMessages` listener

_proposed by: @kabir.kbr; date: 2024-04-17_

TBD, required by `dms.node` interface.

See currently proposed interfaces and data model [mailboxes.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/background_tasks/mailboxes.go).



# Summary
The `background_tasks` package is an internal package responsible for managing background jobs within DMS.
It contains a scheduler that registers tasks and run them according to the schedule defined by the task definition.

## Tasks
Task is a struct that defines a job. It includes the task's ID, Name, the function that is going to be run, the arguments for the function, the triggers that trigger the task to run, retry policy, etc.

### Triggers
Trigger is an interface that defines IsReady and Reset methods. IsReady should return true if the task should be run and Reset resets the trigger until the next event happens.
There are different implementations for the trigger interface.

* PeriodicTrigger: Defines a trigger based on a duration interval or a cron expression.
* EventTrigger: Defines a trigger that is set by a trigger channel.
* OneTimeTrigger: A trigger that is only triggered once after a set delay.

## Scheduler
The sceduler is the orchestrator that manages and runs the tasks.
There is a `NewScheduler` function that creates a new scheduler which takes `maxRunningTasks` argument to limit the maximum number of tasks to run at a time.
If the scheduler task queue is full, remaining tasks that are triggered will wait until there is a slot available in the scheduler.
It has the following functionalities.

* AddTask: Registers a task to be run when triggered.
* RemoveTask: Removes a task from the scheduler. Tasks with only OneTimeTrigger will be removed automatically once run.
* Start: Starts the scheduler to monitor tasks.
* Stop: Stops the scheduler.
