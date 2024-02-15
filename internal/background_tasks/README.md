## Summary
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
