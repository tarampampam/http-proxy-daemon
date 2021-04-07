<p align="center">
  <img src="https://hsto.org/webt/jx/ea/tw/jxeatw6qghfyfzxu2y8cymoiyck.png" alt="Logo" width="128" />
</p>

# Http Proxy Daemon

[![Release version][badge_release_version]][link_releases]
![Project language][badge_language]
[![Build Status][badge_build]][link_build]
[![Release Status][badge_release]][link_build]
[![Coverage][badge_coverage]][link_coverage]
[![License][badge_license]][link_license]

This application allows sending any HTTP requests throughout itself (proxying) using dynamic HTTP route, like `http://app/proxy/https/example.com/file.json?any=param` (request will be sent on `https://example.com/file.json?any=param`). By running this application on a remote server you can send requests to any resources "like from a server" from anywhere!

## Usage example

Run proxy server:

```shell
$ ./http-proxy-daemon serve --port 8080 --prefix 'proxy'
```

Then send an HTTP request to the `https://httpbin.org/get?foo=bar&bar&baz` through our server:

```bash
$ curl -s -H "foo:bar" --user-agent "fake agent" 'http://127.0.0.1:8080/proxy/https/httpbin.org/get?foo=bar&bar&baz'
{
  "args": {
    "bar": "",
    "baz": "",
    "foo": "bar"
  },
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Foo": "bar",
    "Host": "httpbin.org",
    "User-Agent": "fake agent"
  },
  "origin": "8.8.8.8, 1.1.1.1",
  "url": "https://httpbin.org/get?foo=bar&bar&baz"
}
```

## Using docker

[![image stats](https://dockeri.co/image/tarampampam/http-proxy-daemon)][link_docker_tags]

Run docker-container with a proxy server in background _(detached)_ and listen for 8080 TCP port (incoming HTTP requests):

```bash
$ docker run --rm -d -p "8080:8080/tcp" tarampampam/http-proxy-daemon:X.X.X serve --port 8080
```

> Important notice: do **not** use `latest` application tag _(this is bad practice)_. Use versioned tag (like `1.2.3`) instead.

## Benchmark

Start this application in a docker-container:

```bash
$ docker run --rm --net host tarampampam/http-proxy-daemon:0.3.0 serve --port 8080
```

Start `nginx` beside:

```bash
$ docker run --rm --net host nginx:alpine
```

Next, run **Apache Benchmark**:

```bash
$ ab -kc 15 -t 90 'http://127.0.0.1:8080/proxy/http/127.0.0.1:80'
This is ApacheBench, Version 2.3 <$Revision: 1843412 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 5000 requests
Completed 10000 requests
Completed 15000 requests
Completed 20000 requests
Completed 25000 requests
Completed 30000 requests
Completed 35000 requests
Completed 40000 requests
Completed 45000 requests
Completed 50000 requests
Finished 50000 requests


Server Software:        nginx/1.19.9
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /proxy/http/127.0.0.1:80
Document Length:        612 bytes

Concurrency Level:      15
Time taken for tests:   7.469 seconds
Complete requests:      50000
Failed requests:        0
Keep-Alive requests:    50000
Total transferred:      44100000 bytes
HTML transferred:       30600000 bytes
Requests per second:    6694.38 [#/sec] (mean)
Time per request:       2.241 [ms] (mean)
Time per request:       0.149 [ms] (mean, across all concurrent requests)
Transfer rate:          5766.06 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:     0    2  14.1      1     371
Waiting:        0    2  14.1      1     371
Total:          0    2  14.1      1     371

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      2
  75%      2
  80%      2
  90%      3
  95%      5
  98%      6
  99%      7
 100%    371 (longest request)
```

> Hardware info:
> ```shell
> $ cat /proc/cpuinfo | grep 'model name'
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
> model name	: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
>
> $ cat /proc/meminfo | grep 'MemTotal'
> MemTotal:       16261464 kB
> ```


### Supported tags

[![image stats](https://dockeri.co/image/tarampampam/http-proxy-daemon)][link_docker_tags]

All supported image tags [can be found here][link_docker_tags].

### Testing

For application testing we use built-in golang testing feature and `docker-ce` + `docker-compose` as develop environment. So, just write into your terminal after repository cloning:

```shell
$ make test
```

Or build binary file:

```shell
$ make build
```

## Changes log

[![Release date][badge_release_date]][link_releases]
[![Commits since latest release][badge_commits_since_release]][link_commits]

Changes log can be [found here][link_changes_log].

## Support

[![Issues][badge_issues]][link_issues]
[![Issues][badge_pulls]][link_pulls]

If you will find any package errors, please, [make an issue][link_create_issue] in current repository.

## License

This is open-sourced software licensed under the [MIT License][link_license].

[badge_build]:https://img.shields.io/github/workflow/status/tarampampam/http-proxy-daemon/tests?maxAge=30&logo=github
[badge_release]:https://img.shields.io/github/workflow/status/tarampampam/http-proxy-daemon/release?maxAge=30&label=release&logo=github
[badge_coverage]:https://img.shields.io/codecov/c/github/tarampampam/http-proxy-daemon/master.svg?maxAge=30
[badge_release_version]:https://img.shields.io/github/release/tarampampam/http-proxy-daemon.svg?maxAge=30
[badge_language]:https://img.shields.io/github/go-mod/go-version/tarampampam/http-proxy-daemon?longCache=true
[badge_license]:https://img.shields.io/github/license/tarampampam/http-proxy-daemon.svg?longCache=true
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/http-proxy-daemon.svg?maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/http-proxy-daemon/latest.svg?maxAge=45
[badge_issues]:https://img.shields.io/github/issues/tarampampam/http-proxy-daemon.svg?maxAge=45
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/http-proxy-daemon.svg?maxAge=45

[link_coverage]:https://codecov.io/gh/tarampampam/http-proxy-daemon
[link_build]:https://github.com/tarampampam/http-proxy-daemon/actions
[link_docker_hub]:https://hub.docker.com/r/tarampampam/http-proxy-daemon/
[link_docker_tags]:https://hub.docker.com/r/tarampampam/http-proxy-daemon/tags
[link_license]:https://github.com/tarampampam/http-proxy-daemon/blob/master/LICENSE
[link_releases]:https://github.com/tarampampam/http-proxy-daemon/releases
[link_commits]:https://github.com/tarampampam/http-proxy-daemon/commits
[link_changes_log]:https://github.com/tarampampam/http-proxy-daemon/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/http-proxy-daemon/issues
[link_create_issue]:https://github.com/tarampampam/http-proxy-daemon/issues/new/choose
[link_pulls]:https://github.com/tarampampam/http-proxy-daemon/pulls
