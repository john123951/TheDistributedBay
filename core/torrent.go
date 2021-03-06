package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

var Trackers = []string{
	"udp://open.demonii.com:1337/announce",
	"udp://tracker.publicbt.com:80/announce",
	"udp://tracker.istole.it:80/announce",
}

type Torrent struct {
	// Hash of everything in this struct
	Hash string
	// Torrent information
	InfoHash                     []byte
	Name                         string
	Description                  string
	Size                         uint64
	CategoryID                   uint8
	CreatedAt                    time.Time
	Tags                         []string
	Files                        uint
	Seeders, Leechers, Completed Range
}

func (t Torrent) Category() string {
	switch t.CategoryID {
	case 0:
		return "All"
	case 1:
		return "Anime"
	case 2:
		return "Software"
	case 3:
		return "Games"
	case 4:
		return "Adult"
	case 5:
		return "Movies"
	case 6:
		return "Music"
	case 7:
		return "Other"
	case 8:
		return "Series & TV"
	case 9:
		return "Books"
	}
	return "Unknown"
}

func CategoryToId(category string) uint8 {
	switch strings.ToLower(category) {
	case "all":
		return 0
	case "anime":
		return 1
	case "software":
		return 2
	case "games":
		return 3
	case "adult":
		return 4
	case "movies":
		return 5
	case "music":
		return 6
	case "other":
		return 7
	case "series & tv":
		return 8
	case "books":
		return 9
	}
	return 0
}

func (t Torrent) NiceInfoHash() string {
	return hex.EncodeToString(t.InfoHash)
}
func (t Torrent) MagnetLink() string {
	infoHash := t.NiceInfoHash()
	name := url.QueryEscape(t.Name)
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, name)

	for _, tracker := range Trackers {
		magnet += "&tr=" + tracker
	}

	return magnet
}

func CreateTorrent(infoHash []byte, name, description string, category string, createdAt time.Time, tags []string, size uint64, files, seeders, leechers uint) *Torrent {
	category_id := CategoryToId(category)
	t := &Torrent{"", infoHash, name, description, size, category_id, createdAt,
		tags, files, NewRange(seeders), NewRange(leechers), NewRange(0)}
	t.CalculateHash()
	return t
}

func (t *Torrent) CalculateHash() {
	t.Hash = hashTorrent(t)
}

func hashTorrent(t *Torrent) string {
	h := sha256.New()
	io.WriteString(h, (string)(t.InfoHash))
	io.WriteString(h, t.Name)
	io.WriteString(h, t.Description)
	binary.Write(h, binary.LittleEndian, t.CategoryID)
	binary.Write(h, binary.LittleEndian, t.CreatedAt.Unix())
	for _, tag := range t.Tags {
		io.WriteString(h, tag)
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (t *Torrent) VerifyTorrent() error {
	h := hashTorrent(t)
	if h != t.Hash {
		return errors.New(fmt.Sprintf("mutated hash %s vs %s", h, t.Hash))
	}
	return nil
}
