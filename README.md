# POC for a Message Scheduler
This is a bare bones POC application that will read messages from Kafka and schedule them in Redis. There are 2 main services:

* Ingester: ingests messages and storing the timestamp information in Redis via Flink stateful functions (statefun). 
* Metronome: determines which messages to emit on a regular cadence (set to 1 min as a default)

Additionally, the Flink statefun emits to an output topic to record scheduling status.

The architecture is designed to handle high throughput on both the scheduling side and emissions side. Simple benchmarking has shown it can handle scheduling and emitting 100k msgs/sec using the included `docker-compose` setup. 

This project is a POC and NOT battle-hardened for production purposes and would have to be modified to support (among other things):
* Distributed deployment / coordination
* Redis sharding
* Fault tolerance for losing messages upon crashing

Potential TODOs:
* Address the battle-hardening issues
* Add UI to provide real-time scheduling statuses
* Add capability to reschedule or cancel previously scheduled messages
