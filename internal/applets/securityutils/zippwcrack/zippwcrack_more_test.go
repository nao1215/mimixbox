package zippwcrack

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"hash/crc32"
	"testing"
)

// unencryptedZip builds a valid ZIP with a single Store-compressed member and no
// encryption.
func unencryptedZip(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestFirstEncryptedSkipsUnencrypted: an archive with only unencrypted members
// reports that no ZipCrypto entry was found (covering the skip branch).
func TestFirstEncryptedNoEncryptedEntry(t *testing.T) {
	t.Parallel()
	data := unencryptedZip(t, "plain.txt", "no secrets here")
	_, err := firstEncrypted(data)
	if err == nil {
		t.Fatal("expected an error when no encrypted entry exists")
	}
	if err.Error() != "no ZipCrypto-encrypted entry found" {
		t.Errorf("err = %v, want 'no ZipCrypto-encrypted entry found'", err)
	}
}

// TestFirstEncryptedFindsEncrypted confirms the encrypted fixture parses and
// carries the recorded CRC and method.
func TestFirstEncryptedFindsEncrypted(t *testing.T) {
	t.Parallel()
	e, err := firstEncrypted(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(e.encrypted) < 12 {
		t.Errorf("encrypted payload too short: %d bytes", len(e.encrypted))
	}
	if e.crc32 == 0 {
		t.Error("crc32 should be recorded from the archive")
	}
}

// TestVerifyPasswordShortHeader: an entry whose payload is shorter than the
// 12-byte ZipCrypto header can never verify.
func TestVerifyPasswordShortHeader(t *testing.T) {
	t.Parallel()
	e := entry{encrypted: []byte("short"), method: zip.Store}
	if verifyPassword(e, "whatever") {
		t.Error("a sub-header payload must not verify")
	}
}

// TestVerifyPasswordUnsupportedMethod: a member compressed with an unsupported
// method is rejected even with the right key material.
func TestVerifyPasswordUnsupportedMethod(t *testing.T) {
	t.Parallel()
	e := entry{encrypted: make([]byte, 16), method: 99}
	if verifyPassword(e, "x") {
		t.Error("an unsupported compression method must not verify")
	}
}

// TestVerifyPasswordStoreRoundTrip exercises the zip.Store decrypt path
// end-to-end: encrypt a stored payload with a known password, then confirm the
// right password verifies and a wrong one does not.
func TestVerifyPasswordStoreRoundTrip(t *testing.T) {
	t.Parallel()
	const password = "s3cr3t"
	raw := []byte("hello stored world")

	enc := encryptStore(t, raw, password)
	e := entry{
		encrypted: enc,
		crc32:     crc32.ChecksumIEEE(raw),
		method:    zip.Store,
	}

	if !verifyPassword(e, password) {
		t.Error("correct password should verify a Store entry")
	}
	if verifyPassword(e, "wrong") {
		t.Error("wrong password should not verify")
	}
}

// TestVerifyPasswordDeflateRoundTrip exercises the zip.Deflate decode path:
// deflate-compress a body, encrypt it with a known password, and confirm the
// CRC of the inflated plaintext is checked.
func TestVerifyPasswordDeflateRoundTrip(t *testing.T) {
	t.Parallel()
	const password = "deflateKey"
	raw := []byte(bytes.Repeat([]byte("compress me "), 8))

	var comp bytes.Buffer
	fw, err := flate.NewWriter(&comp, flate.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := fw.Close(); err != nil {
		t.Fatal(err)
	}

	enc := encryptStore(t, comp.Bytes(), password)
	e := entry{
		encrypted: enc,
		crc32:     crc32.ChecksumIEEE(raw),
		method:    zip.Deflate,
	}

	if !verifyPassword(e, password) {
		t.Error("correct password should verify a Deflate entry")
	}
	if verifyPassword(e, "nope") {
		t.Error("wrong password should not verify a Deflate entry")
	}
}

// encryptStore produces a 12-byte-header ZipCrypto stream over raw using
// password, mirroring the decrypt cipher in reverse so the fixture is internally
// consistent.
func encryptStore(t *testing.T, raw []byte, password string) []byte {
	t.Helper()
	zc := newZipCrypto(password)
	// A deterministic 12-byte header; its plaintext content is irrelevant to the
	// CRC check, which is computed over the body only.
	header := []byte("ABCDEFGHIJKL")
	plain := append(append([]byte{}, header...), raw...)

	out := make([]byte, len(plain))
	for i, p := range plain {
		out[i] = zc.encryptByte(p)
	}
	return out
}

// encryptByte is the inverse of decryptByte: it XORs the same keystream byte and
// advances the key schedule with the plaintext, exactly as decrypt does.
func (z *zipCrypto) encryptByte(p byte) byte {
	temp := uint16(z.key2|2) & 0xffff
	c := p ^ byte((uint32(temp)*uint32(temp^1))>>8)
	z.update(p)
	return c
}
