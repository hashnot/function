The project can be described as a node-red for dockerised services. Take a look at FAAS providers to get some idea.

Take node-red screen, like this one for example: ![node-red example](http://unipi.technology/wp-content/uploads/screen1.png)
and imagine the left panel lists docker images from a registry. On the main panel, blocks represent containers and the lines represent message queues (AMQP/RabbitMQ).

If you take any two components, a producer and consumer, the communication looks like this:

component A --> exchange --> queue --> component B

(This represents single direction communication but AMQP supports RPC and so will the framework.)

Exchange and queue are part of [AMQP server](https://www.rabbitmq.com/getstarted.html).
Exchange applies some simple logic on the message to decide to which queue to send it.
Queue can be persistent and a single component (can be scaled) consumes messages from it.

The framework will monitor the queue length and scale up or down the component that is processing messages from that queue. It could even scale it down to zero and fire it back up on a message on a preceding queue.

There will be SDKs for java, js and go.
The role of SDK is to handle all framework and AMQP stuff and allow the developers to implement their business logic in a function (or a class) (Function As A Service).
Skala and Akka in particular look very interesting as an integration target.

TODO:
- core functionality - setup AMQP exchanges and queues and fire up docker containers based on a descriptor file. Think of extended docker-compose
- add node-red as UI
- create SKDs for java and js
- port standard components from node-red (where applicable)
- scaling
- limits
- billing and spending limits, if it goes commercial

Links:
- [FAAS in Wikipedia](https://en.wikipedia.org/wiki/Function_as_a_Service)
- [Node-red](https://nodered.org/)
- [IronFunctions](Iron.io) (the first FAAS implementation I could find)
- [Google Cloud Functions](https://cloud.google.com/functions/docs/)
- [Azure Functions](https://azure.microsoft.com/en-us/services/functions/)
- [AWS Lambda](https://aws.amazon.com/lambda/details/)
- [IBM Bluemix OpenWhisk](https://www.ibm.com/cloud-computing/bluemix/openwhisk)
