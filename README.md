# MagicBus

  > Every day I get in the queue
  > -- _The Who, Magic Bus_

## 1. What it is

MagicBus is a _local command/event bus_, which can be extended to a distributed messaging system by inter-connecting remote buses.

The idea is that Aggregates (Domain Driven Design term for subsystem) can simply _"plug in"_ and have their events and commands delivered them in _serialized fashion_.

The bus ***solves the problem of handling concurrency in a concurrent, asynchronous messaging system***.

This occurs in situations where Aggregates have to be able respond to multiple concurrent events or commands.

The _concurrency handling_ is employs the [Actor Model](https://www.youtube.com/watch?v=7erJ1DV_Tlo), which is an evolving [Go pattern](https://www.slideshare.net/weaveworks/an-actor-model-in-go-65174438). The bus itself is an Actor, as is every Aggregate that plugs into the bus.

The bus also implements [Command Query Response Segregation](https://www.youtube.com/watch?v=whCk1Q87_ZI):
- _commands/events_ are handled by the Aggregate directly,
- _queries_ (read model) have to be submitted to a [repository](repository/repository.go),
- Aggregate and Repository share no code or data; the only way the Repository can be updated is via events sent from the Aggregate.


## 2. Implementation status

The bus itself is ready to use - see the test code. For integration, you probably need to fork and customize the generic sketches in the code to your specific needs.

![Magic Bus](https://i.redd.it/wday28h1ruhy.jpg)
