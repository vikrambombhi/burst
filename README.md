# What is burst?
Burst is a pub sub messaging system built in Golang.

# How do I use burst?
1. Start burst `go run main.go`
2. Connect to a topic. This is done by creating a websocket to burst and specifiying the topic in the url for example `ws://localhost:8080/foo` will connect you to topic `foo`
3. Send and Receive messages: This is simply done by sending a message down the websocket. All subscribers will recieve the message including the sender.

# Config
The number of workers a given topic has can be changed with the environment variable `WORKERS_PER_POOL`. The default number of workers per topic is 8.

# API
`/get-topics` this is a reserved topic name. And will return all topics that have been created since burst was started

# Next steps
For now performance is not a issue. The next steps for development will be:
1. Build a client package so the logic of connecting to topics, sending/receiving message can all be abstracted away
1. Benchmarking and testing
1. Implementing logging for messages and other events.
1. Cleaning up closed connections. (Currently a memory leak)
1. Add the abillity to set ports.
