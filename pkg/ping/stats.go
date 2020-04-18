package ping

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// Stats contains the statistics related to a Pinger run
// Pinger uses this struct to aggregate the results and print the statistics at the
// end
type Stats struct {
	// host is the host provided by user to ping
	host string

	// rtts is a slice which contains a list of all the round trip times for the
	// ICMP echo request-reply cycle.
	rtts []time.Duration

	// txPackets contains the number of recieved ICMP packets from the
	// host we tried to ping.
	rxPackets int

	// txPackets contains the number of ICMP packets transmitted.
	txPackets int
}

// NewStats returns a new instance of Stats which can be used for collecting
// ping statistics.
func NewStats(host string) *Stats {
	return &Stats{
		host: host,
		rtts: make([]time.Duration, 0),
	}
}

// IncrementTxPackets increments the number of transmitted packets in the statistics
func (p *Stats) IncrementTxPackets() {
	p.txPackets++
}

// IncrementRxPackets increments the number of recieved packets in the statistics
func (p *Stats) IncrementRxPackets() {
	p.rxPackets++
}

// AddRTT adds a new Round Trip Time to the statistics.
func (p *Stats) AddRTT(t time.Duration) {
	p.rtts = append(p.rtts, t)
}

// Print prints the ping statistics gathered from the Pinger
func (p *Stats) Print() {
	var totalTime, maxTime, minTime time.Duration

	maxTime = time.Nanosecond
	if len(p.rtts) > 0 {
		minTime = p.rtts[0]
	}
	for _, t := range p.rtts {
		totalTime = totalTime + t

		if t > maxTime {
			maxTime = t
		}

		if t < minTime {
			minTime = t
		}
	}

	percentageLoss := ((p.txPackets - p.rxPackets) / p.txPackets) * 100

	log.Infof("--- %s ping statistics ---", p.host)
	log.Infof("%d packets transmitted, %d recieved, %d percent packet loss, time %s", p.txPackets, p.rxPackets, percentageLoss, totalTime)
	log.Infof("rtt min/avg/max = %s/%fms/%s", minTime, float64(totalTime.Milliseconds())/float64(len(p.rtts)), maxTime)
}
