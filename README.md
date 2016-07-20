# youtrend

youtrend is a program that is designed to connect to redis queues and read the urls in under the www.youtube.com_crawler_queue key
and parse video details from the video page and then calculate the zscore of each video and insert the results in a collection in
mongodb.

This allows for easily querying videos that are above a certain z-score which may in-turn be trending on youtube. More statistics
should be add to the VideoStats struct so that we can then query on more than just a z-score. Still looking into what may be good
features to add to the struct which is persisted in the mongo collection.

## Dependencies

This is meant to be used along side with go-crawler which is a web-crawler that only crawls pages that match to a certain regex.
So we can use this to specify only youtube videos by using a valid regex for youtube videos. Since go-crawler dumps the urls into
a redis queue with the domain name of the site prepended to `_crawler_queue`, we can then run this program after we have crawler
a certain depth of youtube or more boldly all videos on youtube or those who look like keen candidates for those to be trending.

## Future uses

Once more data can be extracted by the site it may prove useful to begin storing all this data so that machine learning techniques
can be applied on the data to better determine which videos may be likely to gain popularity (views), subscribers.

## Setup

Setting up youtrend is as simple as running the `bootstrap.sh` script which will ensure that Go is installed and that your `$GOPATH`
is set up. Then install the project dependencies, and attempt to build the binary in the root project folder so that it may be easily
added when creating the Docker image.

```
$ ./bootstrap.sh
```

## Docker & Docker compose

Currently this project is to be used in conjunction with Docker and docker-compose for small-scale testing as more of a proof of
concept. While it could be easily used with production redis and mongodb instances by means of manipulating environment variables
at runtime, it has still yet to be tested.

Build the docker image:

```
$ docker build -t youtrend:latest .
```

Or using docker-compose:

```
$ docker-compose build
```

## Running youtrend

As stated before running this program with production instances of redis and mongodb has yet to be tested. Yet, I do not see why
they could not be used so long as the host and port of redis and mongodb are specified in the proper environment variables.

| ENVIRONMENT VARIABLE NAME   |   ENVIRONMENT VARIABLE VALUE      |
|-----------------------------|:---------------------------------:|
|  REDIS_PORT_6379_TCP_ADDR   | The ip address of redis instance  |
|  REDIS_PORT_6379_TCP_PORT   | Redis server port (default 6739)  |
| MONGODB_PORT_27017_TCP_ADDR | The ip address of mongo instance  |
| MONGODB_PORT_27017_TCP_PORT | Mongo server port (default 28017) |

Locally:

```
$ REDIS_PORT_6379_TCP_ADDR=your_redis_server_ip
$ REDIS_PORT_6379_TCP_PORT=your_redis_server_port
$ MONGODB_PORT_27017_TCP_ADDR=your_mongodb_server_ip
$ MONGODB_PORT_27017_TCP_PORT=your_mongodb_server_port
$ ./youtrend
```

Docker compose:

```
$ docker-compose up
```

## License

MIT License
