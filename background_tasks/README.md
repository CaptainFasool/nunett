## Introduction

This package is for scheduling background tasks to preserve resources. Takes care of background tasks scheduling and execution, other packages that have their own background tasks register through this package:
1. Registration 
    1. The task itself, the arguments it needs
    2. priority 
    3. event (time period or other event to trigger task)
2. Start , Stop, Resume
3. Algorithm that accounts for the event and priority of the task (not yet clear) 
4. Monitor resource usage of tasks (not yet clear)

(Definition to be refined)