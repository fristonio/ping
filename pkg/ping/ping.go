package ping

import (
	"fmt"
	"math/rand"
	"net"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Pinger is a type which sends and recieves ICMP echo packets
// to replicate ping functionalities.
type Pinger struct {
	// ipaddr is the IP address of the host.
	ipaddr *net.IPAddr

	isIPV6 bool

	// Size is the size of the payload to send in the ICMP echo packets
	Size int

	// identifier is the identifier associated with the ICMP packet
	identifier int

	// seqNum is the sequence number of the ICMP packet to be sent.
	seqNum int

	// shutdown is the channel which is used to shut down the running pinger
	// instance.
	shutdown chan bool

	// conn represents the ICMP connection to the remote host.
	conn *icmp.PacketConn

	// maxRTT is the maximum Round Trip Time for a packet.
	// If a packet reply is recieved after this time then it is discarded.
	maxRTT time.Duration

	// timer is the time used for calculation of RTT for the ICMP echo request
	// reply cycle.
	timer time.Time

	stats *Stats
}

// NewPinger returns a new instance of Pinger with the host corresponding to the
// provided host address.
func NewPinger(addr string) (*Pinger, error) {
	var ipaddr net.IP
	ipaddr = net.ParseIP(addr)
	if ipaddr == nil {
		ipaddrs, err := net.LookupIP(addr)
		log.Debugf("%v", ipaddrs)
		if err == nil && len(ipaddrs) > 0 {
			ipaddr = ipaddrs[0]
		} else {
			return nil, fmt.Errorf("error while looking up ip for provided host")
		}
	}

	return &Pinger{
		ipaddr:     &net.IPAddr{IP: ipaddr},
		isIPV6:     len(ipaddr.To4()) != net.IPv4len,
		identifier: rand.Intn(0xffff),
		seqNum:     0,
		shutdown:   make(chan bool),
		maxRTT:     time.Second,
		timer:      time.Now(),
		stats:      NewStats(addr),
	}, nil
}

// setupConnection starts running the ICMP packets listner.
func (p *Pinger) setupConnection() error {
	log.Debug("setting up ICMP connection for pinger")
	protocol := "ip4:icmp"
	if p.isIPV6 {
		protocol = "ip6:ipv6-icmp"
	}

	conn, err := icmp.ListenPacket(protocol, "")
	if err != nil || conn == nil {
		return fmt.Errorf("error while starting listner: %s", err)
	}

	p.conn = conn
	return nil
}

func (p *Pinger) sendIcmp() error {
	var icmpType icmp.Type
	icmpType = ipv4.ICMPTypeEcho
	if p.isIPV6 {
		icmpType = ipv6.ICMPTypeEchoRequest
	}

	p.timer = time.Now()
	data, err := (&icmp.Message{
		Type: icmpType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.identifier,
			Seq:  p.seqNum,
			Data: make([]byte, 64),
		},
	}).Marshal(nil)

	if err != nil {
		return fmt.Errorf("error while marshling icmp packet: %s", err)
	}

	go func(conn *icmp.PacketConn, addr net.Addr, data []byte) {
		log.Debug("sending ICMP packet to the host.")
		for {
			if _, err := conn.WriteTo(data, addr); err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Err == syscall.ENOBUFS {
						continue
					}
				}
			}
			break
		}

		p.stats.IncrementTxPackets()
	}(p.conn, p.ipaddr, data)

	return nil
}

func (p *Pinger) recvIcmp() error {
	for {
		select {
		case <-p.shutdown:
			return nil
		default:
			// create a byte slice big enough to contain the packet.
			data := make([]byte, 1024)
			len, addr, err := p.conn.ReadFrom(data)

			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						log.Warn("read timeout for ICMP packet.")
						continue
					} else {
						log.Warn("error while reading ICMP packet from connection")
					}
				}
			}

			// We have recieved the packet, validate the packet and print the information
			// regarding ICMP echo reply.
			log.Debugf("recieved ICMP packet from host: %s: len: %d", addr, len)
			proto := 1
			if p.isIPV6 {
				proto = 58
			}

			message, err := icmp.ParseMessage(proto, data)
			if err != nil {
				return fmt.Errorf("error while parsing icmp message: %s", err)
			}

			if message.Type != ipv4.ICMPTypeEchoReply &&
				message.Type != ipv6.ICMPTypeEchoReply {
				log.Debugf("icmp message is not an Echo reply: %s", message.Type)
				continue
			}

			var rtt time.Duration
			switch packet := message.Body.(type) {
			case *icmp.Echo:
				if packet.ID == p.identifier && packet.Seq == p.seqNum {
					rtt = time.Since(p.timer)

					log.Infof("%d bytes from %s: icmp_seq=%d ttl=%d time=%s",
						len, p.ipaddr.String(), p.seqNum, 53, rtt)
					p.stats.AddRTT(rtt)
					p.stats.IncrementRxPackets()
					return nil
				}
			default:
				log.Debug("recieved ICMP message is not of Echo type")
			}
		}
	}
}

// Run starts running pinger to ping the configured host.
func (p *Pinger) Run() error {
	log.Debug("starting to run pinger")
	err := p.setupConnection()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(p.maxRTT)
	for {
		select {
		case <-p.shutdown:
			ticker.Stop()
			p.conn.Close()
			return nil
		case <-ticker.C:
			err := p.sendIcmp()
			if err != nil {
				log.Errorf("error while sending ICMP packet: %s", err)
			} else {
				err = p.recvIcmp()
				if err != nil {
					log.Errorf("error while recieving ICMP echo reply: %s", err)
				}
			}

			p.seqNum++
		}
	}
}

// Shutdown shuts down the running pinger instance.
func (p *Pinger) Shutdown() {
	p.shutdown <- true
}

// PrintStats prints the statistics aggregated the pinger.
func (p *Pinger) PrintStats() {
	p.stats.Print()
}
