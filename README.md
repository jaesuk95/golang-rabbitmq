This service receives IoT data via HTTP POST and publishes it to RabbitMQ queues for asynchronous processing and analysis. It is built with Go and uses `streadway/amqp` for AMQP communication.

---

- Accepts HTTP POST requests on:
  - `/api/iot/car` → Publishes to `iot-data` queue
  - `/api/iot/charge` → Publishes to `iot-data-charge` queue
- JSON-based payloads for IoT data
- Designed for lightweight ingestion and high-throughput scenarios
- Easily extensible to support new IoT data types and queues

--- 
RabbitMQ client library for Go:
go get github.com/streadway/amqp
