# Broadcasting messages in Go with Ably and Redis

This project contains the fundamentals for a fanout-based system using Ably. Redis is used as a queueing system to hold messages until they're ready to be sent by Ably. This demo has the ability to ensure that any number of producers of data and publishers of data don't exceed rate limits within Ably, ensuring a reliable and consistent amount of traffic to any subscribed clients.

More in depth details of this can be found on the associated blog post on the Ably Blog.

## Setup

Fundamentally to run this all you need is a Kubernetes Cluster on which to apply the contents of the `deployments` folder. You'll need to ensure you've got a free [Ably Account](https://www.ably.com/signup), and then replace the `ABLY_KEY` environment variable inside of `ably-publisher-deployment.yml` with your own [Ably API key](https://ably.com/accounts/any/apps/any/app_keys).

Once you've done that, you can apply the Kubernetes configuration files to your cluster. If you don't have one, an easy one to get started with locally is [minikube](https://minikube.sigs.k8s.io/docs/start/). You'll also need [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/) to run Kubernetes commands.

To apply the configuration files, run the following:

```sh
> kubectl apply -f redis.yml
> kubectl apply -f ably-publisher-deployment.yml
> kubectl apply -f trading-deployment.yml
```

To delete the deployments, you'll just need to run the following:

```sh
> kubectl delete deployment redis-deployment
> kubectl delete deployment ably-publisher
> kubectl delete deployment trading-server
```

You can view messages being sent from your setup by loading the html page `client.html` from inside the `client` folder.

## Details

The contents of the kubernetes cluster are generated from the `trading-server` folder and `ably-publishing-server` folders.

* *trading-server* contains a Go program which generates random data to represent new trade values, and publishes it to a Redis queue
* *ably-publishing-server* contains a Go program which consumes from the Redis queue, and publishes the data into Ably. All instances of this program will take into consideration a rate limit which is determined by a rolling log contained within Redis. This ensures the publish rate into Ably never exceeds the rate permitted.

These programs are then turned into Docker Images using their relevant Dockerfile, and deployed to Docker Hub for usage by the Kubernetes configuration files. The Image for the trading server can be found at https://hub.docker.com/repository/docker/tomably/trading-server, and the publishing server at https://hub.docker.com/repository/docker/tomably/go-redis-ably-publisher.
