# docker-log-watch

Simple tool to follow docker container logs on console.

I was dissatisfied with the way the build-in `docker-compose logs -f` handles new started containers
( for some unknown reason it didn't pick up on restarted containers ? ).

To get started with `go` I simply used the original go wrapper of `docker` in [Develop with Docker Engine SDKs](https://docs.docker.com/engine/api/sdk/)
and wrote this little console tool.

`docker-log-watch` will simply output the log lines of all running docker containers and print them
with a little colorized prefix in your console.

