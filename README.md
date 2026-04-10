# kapbit-go

Kapbit is a lightweight, high-performance workflow orchestrator for Go, designed 
specifically for the Saga pattern. It provides a robust framework for building 
long-running, fault-tolerant workflows without the overhead of heavy infrastructure.

## Why Kapbit?

- **Log-Oriented Architecture**: Unlike traditional orchestrators that rely on
  distributed databases, Kapbit is built on a distributed log (currently
  supporting Kafka). This offers superior performance, natural event sourcing,
  and simplified data consistency.
- **No DSL Required**: No custom Domain Specific Languages or complex JSON/YAML
  definitions. If you know Go, you know Kapbit.
- **Built for Simplicity**: A minimalist API that stays out of your way, letting
  you focus on business logic rather than distributed transaction plumbing.

## Key Features

- **Fencing**: Guarantees a single active writer, preventing split-brain
  scenarios.
- **Horizontal Scalability**: Distributes load across partitions to scale
  throughput.
- **Fault Tolerance**: Automatically reconnects to storage and resumes workflow
  execution after restarts.
- **Workflow Idempotency**: Processes each unique workflow exactly once.
- **Circuit Breaker Support**: Provided for both storage (built-in) and workflow
  (user-managed) layers.
- **Extensible Design**: Built to support various storage backends and codecs;
  currently ships with Kafka and JSON support.
  

## Examples

For complete usage examples, visit the [examples-go](https://github.com/kapbit/examples-go) repository.

## How It Works

Let's look at the Kapbit components.

### Kapbit Instance

Before launching any new workflow, a Kapbit Instance must establish itself as
an authorized Writer for the storage. It does so by emitting an Active Writer
event to all available partitions. The storage layer ensures only one instance
can hold writer status at a time, preventing conflicting writes and split-brain
scenarios.

Once authority is confirmed, the instance enters a recovery phase by loading
recent events from the storage to populate the Idempotency Window (with Workflow
IDs) and resume the execution of uncompleted workflows.

After initialization, Kapbit is ready to accept new workflows until the
`MaxWorkflows` threshold is met - accounting for both newly launched and
resumed workflows.

### Workflow

A workflow starts with a Workflow Created event being written to the log. No
business logic runs until this event is successfully saved. Once persisted, the
execution is guaranteed - if the system crashes right after saving the event,
the new instance recovery process will find the record and resume the workflow.

In Kapbit, a workflow is a state-managed pipeline. Its structure and data flow
are defined by the following components:

- Workflow Type: Defines a specific sequence of steps to execute.
- Step Actions: Each step consists of an execution and an optional
  compensation action.
- Step Outcome: An action return value represents a step's outcome.
- Progress State: Outcomes are accumulated into a Progress state, which is
  passed forward to each subsequent step and Result Builder.
- Result Builder: Produces the final workflow result.

The diagram below shows the successful execution (happy path).

```
        ----------                   --------
       | Workflow |                 | Result |
        ----------                   --------
            |                            ^
   Progress |                            |
            v                            |
         -------                         |
        | Step1 |                        |
        |  exec |                        |
         -------                         |
            |                            |
   Progress |                            |
            v                            |
         -------                  ---------------
        | Step2 |  ----------->  | ResultBuilder |
        |  exec |    Progress     ---------------
         -------
```

When a step returns a failure outcome, the workflow stops moving forward.
Instead, it "backtracks" and runs the compensation for every step that already
finished.

The diagram below shows Step2 failing and triggering the cleanup for Step1.

```
        ----------               --------
       | Workflow |             | Result |
        ----------               --------
            |                        ^
   Progress |                        |
            v                        |
         -------                     |
        | Step1 |             ---------------
        |  exec |            | ResultBuilder |
         -------              ---------------
            |                        ^
   Progress |                        | Progress
            v                        |
         -------                  -------
        | Step2 |  ----------->  | Step1 |
        |  exec |    Progress    |  comp |
         -------                  -------
```

### Circuit Breakers

Kapbit supports two types of Circuit Breakers:
- **System Level (Storage)**: A built-in circuit breaker for the storage layer.
- **Workflow Level (External Services)**: User-defined circuit breakers for
  remote services and third-party APIs. Returning a `CircuitBreakerOpenError`
  from a workflow step signals to Kapbit that a downstream dependency is
  unavailable.

### Retry Worker

A workflow can fail (return a user-defined error) for only one reason: the 
remote service it depends on is unavailable. In all other cases it should return
a result, even after the compensation phase.

The Retry Worker runs in the background to re-attempt failed workflows,
resuming from the last failed step. It operates in two modes:
- **Fast Mode (Default)**: Handles temporary issues (like network blinks or
  timeouts) with immediate or high-frequency retries.
- **Slow Mode**: Triggered when frequent retries make no sense, for example,
  when a Circuit Breaker opens.

If a workflow exceeds its maximum retry limit, the worker terminates it with a
Dead Letter event for manual handling.

### Events and Emitter

Different system components emit different events:
- Kapbit Instance: Emits Active Writer, Workflow Created and Rejected events.
- Workflow: Emits Step Outcome and Workflow Result events.
- Retry Worker: Emits Dead Letter and Rejected events.

Where Rejected event is used to terminate the workflow if some of its event was
rejected by the storage, or there was an encoding error.

All components delegate event emission to the Emitter, which retries 
indefinitely using exponential backoff. This ensures that events are eventually 
processed even during temporary outages.

While an alternative approach might be to give up and emit a Dead Letter after
several failed attempts, this wouldn't work - if the storage itself is
unavailable the Dead Letter saving would also fail. Therefore, retrying 
indefinitely is the only safe strategy.

### Observability & Async Results

Kapbit is an execution engine, not a query service. It does not include a 
built-in component for retrieving asynchronous workflow results or historical 
data. 

Following the **CQRS (Command Query Responsibility Segregation)** pattern, the 
log storage serves as the **Single Source of Truth**. To implement features 
like a results dashboard or status API, you should:
- **Consume and filter** events directly from the log storage.
- **Project** those events into one or more read-optimized databases (e.g., 
  Key/Value stores, relational DBs, or search indexes).

This decoupled approach ensures that the execution engine remains lightweight 
and highly performant, while giving you the flexibility to build multiple 
specialized views of your data.

### Ingress Gate

When a Circuit Breaker opens, the system can no longer function properly and
should stop accepting new workflows. This is exactly what the Ingress Gate is
for. Both the Kapbit Instance and the Emitter close it when they encounter a
`CircuitBreakerOpenError`. 

The gate becomes open again, only after successful workflow execution, processed
by Kapbit Instance or Retry Worker. During the close period, the last one can 
use already failed workflows to probe the system and reopen the gate as soon as 
possible.

### Repository And Fencing

The Repository provides an abstraction layer over the storage. It is designed 
for high availability and automatically reconnects if the connection drops.

The Repository also provides fencing guarantees. Fencing ensures that only the 
active Kapbit instance can modify the storage. If a write or connection attempt 
returns a Fenced error, it means the instance has lost ownership and can no 
longer modify the storage. In this case, it will immediately terminate.

### Codec

The Repository depends on a user-provided Codec, to encode the Workflow related 
data, such as: input, outcomes, result. At the moment only the JSON format is 
supported.
