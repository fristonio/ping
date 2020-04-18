# Ping

Ping is a simple go implementation of linux ping utility. This project is done as a part of 
cloudflare systems internship assignment.

The implemented version of ping supports both IPv4 and IPv6 along with configurable option for 
setting TTL of the packets.

To run the project follow the below setps.

```sh
$ git clone https://github.com/fristonio/ping

$ make build

[*] Building ping
[*] Setting capabilities for the binary
[*] Build Complete.

$ ./ping -h google.com

INFO[0001] 72 bytes from 216.58.200.174: icmp_seq=0 ttl=53 time=6.706585ms 
INFO[0002] 72 bytes from 216.58.200.174: icmp_seq=1 ttl=53 time=6.649031ms 
INFO[0003] 72 bytes from 216.58.200.174: icmp_seq=2 ttl=53 time=6.856074ms 
INFO[0003] Signal recieved, shutting down pinger instance. 
INFO[0003] --- google.com ping statistics ---           
INFO[0003] 3 packets transmitted, 3 recieved, 0 percent packet loss, time 20.21169ms 
INFO[0003] rtt min/avg/max = 6.649031ms/6.666667ms/6.856074ms
```