// Package zippwcrack implements the zip-pwcrack applet: recover the password of
// a ZIP archive encrypted with traditional PKWARE (ZipCrypto) encryption by
// trying each word of a wordlist, for authorized recovery and testing.
//
// This is a clean-room implementation written from the documented PKWARE
// APPNOTE traditional-encryption algorithm; no third-party cracker source is
// copied. archive/zip is used only to parse the (unencrypted) ZIP structure.
package zippwcrack

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"hash/crc32"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the zip-pwcrack applet.
type Command struct{}

// New returns a zip-pwcrack command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "zip-pwcrack" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Recover the password of a ZipCrypto-encrypted archive"
}

// Run executes zip-pwcrack.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "ARCHIVE -w WORDLIST", stdio.Err)
	wordlist := fs.StringP("wordlist", "w", "", "file of candidate passwords, one per line")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *wordlist == "" {
		return command.Failuref("a wordlist is required (-w)")
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return command.Failuref("exactly one ARCHIVE is required")
	}

	data, err := os.ReadFile(rest[0]) //nolint:gosec // operating on the user-named archive is the point
	if err != nil {
		return command.Failuref("%s", command.FileError(rest[0], err))
	}

	entry, err := firstEncrypted(data)
	if err != nil {
		return command.Failuref("%v", err)
	}

	words, err := loadWords(stdio, *wordlist)
	if err != nil {
		return err
	}

	for _, w := range words {
		if verifyPassword(entry, w) {
			_, _ = io.WriteString(stdio.Out, "password found: "+w+"\n")
			return nil
		}
	}
	_, _ = io.WriteString(stdio.Out, "password not found in wordlist\n")
	return &command.ExitError{Code: command.ExitFailure}
}

// entry holds the raw encrypted bytes of one archive member plus the metadata
// needed to verify a candidate password.
type entry struct {
	encrypted []byte // 12-byte ZipCrypto header followed by the compressed data
	crc32     uint32
	method    uint16
}

// firstEncrypted finds the first ZipCrypto-encrypted member of the archive.
func firstEncrypted(data []byte) (entry, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return entry{}, err
	}
	for _, f := range zr.File {
		if f.Flags&0x1 == 0 {
			continue // not encrypted
		}
		off, err := f.DataOffset()
		if err != nil {
			return entry{}, err
		}
		end := off + int64(f.CompressedSize64)
		if off < 0 || end > int64(len(data)) {
			return entry{}, errors.New("archive entry data is out of bounds")
		}
		return entry{
			encrypted: data[off:end],
			crc32:     f.CRC32,
			method:    f.Method,
		}, nil
	}
	return entry{}, errors.New("no ZipCrypto-encrypted entry found")
}

// verifyPassword reports whether password decrypts the entry to data whose
// CRC32 matches the archive's recorded checksum (an authoritative check, so
// there are no false positives).
func verifyPassword(e entry, password string) bool {
	if len(e.encrypted) < 12 {
		return false
	}
	zc := newZipCrypto(password)
	plain := make([]byte, len(e.encrypted))
	for i, b := range e.encrypted {
		plain[i] = zc.decryptByte(b)
	}
	compressed := plain[12:] // first 12 bytes are the encryption header

	var raw []byte
	var err error
	switch e.method {
	case zip.Store:
		raw = compressed
	case zip.Deflate:
		fr := flate.NewReader(bytes.NewReader(compressed))
		raw, err = io.ReadAll(fr)
		_ = fr.Close()
		if err != nil {
			return false
		}
	default:
		return false
	}
	return crc32.ChecksumIEEE(raw) == e.crc32
}

// zipCrypto implements the PKWARE traditional stream cipher (three rolling
// 32-bit keys seeded from the password).
type zipCrypto struct {
	key0, key1, key2 uint32
}

// crcTab is the standard IEEE CRC-32 table the cipher's key schedule uses.
var crcTab = crc32.MakeTable(crc32.IEEE)

// newZipCrypto seeds the cipher with password.
func newZipCrypto(password string) *zipCrypto {
	zc := &zipCrypto{key0: 0x12345678, key1: 0x23456789, key2: 0x34567890}
	for i := 0; i < len(password); i++ {
		zc.update(password[i])
	}
	return zc
}

// update advances the three keys by one plaintext byte.
func (z *zipCrypto) update(c byte) {
	z.key0 = crc32Update(z.key0, c)
	z.key1 = (z.key1 + (z.key0 & 0xff)) * 134775813 + 1
	z.key2 = crc32Update(z.key2, byte(z.key1>>24))
}

// decryptByte decrypts one ciphertext byte and advances the key schedule.
func (z *zipCrypto) decryptByte(c byte) byte {
	temp := uint16(z.key2|2) & 0xffff
	plain := c ^ byte((uint32(temp)*uint32(temp^1))>>8)
	z.update(plain)
	return plain
}

// crc32Update folds one byte into a running CRC the way the cipher specifies.
func crc32Update(crc uint32, b byte) uint32 {
	return crcTab[(crc^uint32(b))&0xff] ^ (crc >> 8)
}

// loadWords reads the wordlist into a slice.
func loadWords(stdio command.IO, path string) ([]string, error) {
	r, err := command.Open(stdio, path)
	if err != nil {
		return nil, command.Failuref("%s", command.FileError(path, err))
	}
	defer func() { _ = r.Close() }()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, command.Failure(err)
	}
	var words []string
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimRight(line, "\r")
		if len(line) > 0 {
			words = append(words, string(line))
		}
	}
	return words, nil
}
