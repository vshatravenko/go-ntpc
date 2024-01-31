package ntp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"time"
)

var (
	ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	epochDiff       = 2208988800 // difference in seconds between UNIX and NTP epochs(70 years)
	bufSize         = 512
	leapNoIndicator = 0
	clientMode      = 3
	ntpVersion      = 4
	nanoPerSec      = 1000000000
)

/*
ntpTime is a timestamp consisting of two 32 bit ints denoting the number of seconds and fractions respectively
fractions are 1/(2^32) of a second(~2 nanoseconds)
*/
type ntpTime uint64

func (t ntpTime) Duration() time.Duration {
	sec := (t >> 32) * nanoPerSec
	frac := (t & 0xffffffff) * nanoPerSec
	nanoSec := frac >> 32
	if uint32(frac) >= 0x80000000 { // if the unsigned representation is bigger than int32 bounds
		nanoSec++
	}

	return time.Duration(sec + nanoSec)
}

func (t ntpTime) Time() time.Time {
	return ntpEpoch.Add(t.Duration())
}

// ntpTimeShort is similar to ntpTime, but it houses two 16-bit ints instead
type ntpTimeShort uint32

func (t ntpTimeShort) Duration() time.Duration {
	sec := uint64(t>>16) * nanoPerSec
	frac := uint64(t&0xffff) * nanoPerSec
	nanoSec := frac >> 16
	if uint16(nanoSec) >= 0x8000 { // if the unsigned representation is bigger than int16 bounds
		nanoSec++
	}

	return time.Duration(sec + nanoSec)
}

func (t ntpTimeShort) Time() time.Time {
	return ntpEpoch.Add(t.Duration())
}

func toNtpTime(t time.Time) ntpTime {
	nanoSec := uint64(t.Sub(ntpEpoch))
	sec := nanoSec / nanoPerSec
	nanoSecFrac := uint64(nanoSec-sec*nanoPerSec) << 32
	frac := (nanoSecFrac) / nanoPerSec
	if nanoSecFrac%nanoPerSec >= nanoPerSec/2 { // round up the fraction count
		frac++
	}

	return ntpTime(sec<<32 | frac)
}

type mode uint8

/*

Valid packet: [35 0 0 32 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 82 173 112 115 162 69 196 53]
		   0                   1                   2                   3
	       0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |LI | VN  |Mode |    Stratum     |     Poll      |  Precision   |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

type packet struct {
	LIVNMode       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      ntpTimeShort
	RootDispersion ntpTimeShort
	RefID          uint32
	RefTS          ntpTime
	OriginTS       ntpTime
	ReceiveTS      ntpTime
	TransmitTS     ntpTime
}

func initPacket() *packet {
	h := new(packet)

	h.setLI(leapNoIndicator)
	h.setMode(clientMode)
	h.setVersion(ntpVersion)
	h.Precision = 0x20
	h.TransmitTS = randTS()

	return h
}

func packetFromBytes(buf []byte) *packet {
	res := new(packet)
	bReader := bytes.NewReader(buf)
	binary.Read(bReader, binary.BigEndian, res)

	return res
}

/*
ta = T(B) - T(A) = 1/2 * [(T2-T1) + (T3-T4)]
re T1-4 are the four most recent timestamps in order

t is the time of receival on client side(because it's the response destination)
*/
func (h *packet) offset(dest time.Time) time.Duration {
	a := h.ReceiveTS.Time().Sub(h.OriginTS.Time())
	b := h.TransmitTS.Time().Sub(dest)

	return (a + b) / time.Duration(2)
}

/*
ta = T(ABA) = (T4-T1) - (T3-T2).
*/
func (h *packet) rtDelay(recv time.Time) time.Duration {
	a := recv.Sub(h.OriginTS.Time())
	b := h.TransmitTS.Time().Sub(h.ReceiveTS.Time())

	return a - b
}

/*
Causality errors occur when the send timestamp is greater than receive timestamp
The *greater* of the past two causality error is the mininimal error
*/
func (h *packet) minError(dest ntpTime) time.Duration {
	var err0, err1 ntpTime

	if h.OriginTS >= h.ReceiveTS {
		err0 = h.OriginTS - h.ReceiveTS
	}

	if h.TransmitTS >= dest {
		err1 = h.TransmitTS - dest
	}

	return max(err0, err1).Duration()
}

func (h *packet) setLI(li int) {
	h.LIVNMode = (h.LIVNMode & 0x3f) | uint8(li)<<6
}

func (h *packet) setVersion(v int) {
	h.LIVNMode = (h.LIVNMode & 0xc7) | uint8(v)<<3 // 0xc7 == 0b11000111
}

func (h *packet) setMode(m int) {
	h.LIVNMode = (h.LIVNMode & 0xf8) | uint8(m)
}

func (h *packet) getLI() uint8 {
	return (h.LIVNMode ^ 0x3f) >> 6
}

func (h *packet) toBytes() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, h)

	return buf.Bytes()
}

func exponentInterval(n int8) time.Duration {
	// Go doesn't have abs :o
	switch {
	case n > 0:
		return time.Duration(uint64(time.Second) << uint(n))
	case n < 0:
		return time.Duration(uint64(time.Second) << uint(-n))
	default:
		return time.Second
	}
}

func randTS() ntpTime {
	bits := make([]byte, 8)
	rand.Read(bits)

	return ntpTime(binary.BigEndian.Uint64(bits))
}
