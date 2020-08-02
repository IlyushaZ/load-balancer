## Round robin load balancer
This application allows you to fan out your 
http requests to different backends via 
round robin - the easiest strategy of load balancing. 

#### Usage:
- Pass IPs (or domain names) to ```backends.txt``` 
(one address per line)
- Build app via Docker (Dockerfile included) or build binary by running
```go build``` and passing ```-f``` argument with path to file
with list of backends (by default it is ```backends.txt``` in a current directory).
You can also choose the schema (http/https) via ```-s``` argument. 

**Notice:** you can build demo version of program by running ```./build.sh```. 
It requires Docker to be installed.

#### Features: 
- Round robin load balancing
- No limitations for the amount of backends to be used
- Health-checks for the backends every 3 minutes

#### TODO: 
- Add an ability to make several attempts to access the 
backend if no good response received
- Add unit tests
- Add some other strategies of balancing