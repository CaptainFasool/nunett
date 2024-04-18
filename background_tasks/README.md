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