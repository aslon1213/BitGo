package torrentfile

import (
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

type BencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type BencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     BencodeInfo `bencode:"info"`
}

func Open(r io.Reader) (*BencodeTorrent, error) {
	ben := BencodeTorrent{}
	err := bencode.Unmarshal(r, &ben)
	if err != nil {
		return nil, err
	}
	return &ben, nil
}

func (t *BencodeTorrent) ToTorrentFile() (*TorrentFile, error) {
	// TODO: check if infohash is correct
	return nil, nil
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *TorrentFile) BuildTrackerUrl(peerId [20]byte, port int) (string, error) {

	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"comapact":   []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil

}

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

func Unmarshal(peersBin []byte) ([]Peer, error) {
	const Peersize = 6
	if len(peersBin)%Peersize != 0 {
		return nil, errors.New("invalid peers binary")
	}
	number_of_peers := len(peersBin) / Peersize
	peers := make([]Peer, number_of_peers)

	for i := 0; i < number_of_peers; i++ {
		offset := i * Peersize
		IP := net.IP(peersBin[offset : offset+4])
		port := uint16(peersBin[offset+4])<<8 + uint16(peersBin[offset+5])
		peers[i] = Peer{IP, port}
	}
	return peers, nil
}

type TcpHandshake struct {
	Pstr     string
	Infohash [20]byte
	PeerID   [20]byte
}

// Downloading a torrent file from a peer
// Start a TCP connection with the peer. This is like starting a phone call.
// Complete a two-way BitTorrent handshake. “Hello?” “Hello."
// Exchange messages to download pieces. “I’d like piece #231 please."

// func Download(peer Peer) {
// 	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
// 	if err != nil {
// 		return
// 	}

// }

// Serialize serializes the handshake to a buffer
func (h *TcpHandshake) Serialize() []byte {
	buf := make([]byte, 49+len(h.Pstr))
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], h.Infohash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

func Read(r io.Reader) (*TcpHandshake, error) {
	// serialize but backwards
	// read the first byte
	h := TcpHandshake{}
	buf := []byte{}
	_, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	cur := 1
	h.Pstr = string(buf[cur : cur+int(buf[0])])
	cur += int(buf[0])
	// read the 8 bytes
	cur += 8
	// read the infohash
	cur += copy(h.Infohash[:], buf[cur:cur+20])
	cur += copy(h.PeerID[:], buf[cur:cur+20])
	return &h, nil
}
