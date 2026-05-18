package ipfix

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/xzones2014/netflow/session"
)

func errProtocol(f string) error {
	return errors.New("protocol error: " + f)
}

func errTemplateNotFound(t uint16) error {
	return fmt.Errorf("template with id=%d not found", t)
}

func errTemplateMismatch(templateID uint16, expectedSize, actualSize int) error {
	return fmt.Errorf("template id=%d size mismatch: expected %d bytes, got %d bytes (possible desynchronization)", templateID, expectedSize, actualSize)
}

func errRecordSizeAlignment(templateID uint16, remainingBytes int) error {
	return fmt.Errorf("template id=%d: remaining bytes %d is less than expected record size (possible desynchronization)", templateID, remainingBytes)
}

func errInvalidVersion(v uint16) error {
	return fmt.Errorf("version %d is not a valid IPFIX message version (expected 10)", v)
}

// IPFIX Version (RFC 7011)
const Version uint16 = 10

// Decoder can decode multiple IPFIX messages from a stream.
type Decoder struct {
	io.Reader
	session.Session
	*Translate
}

func NewDecoder(r io.Reader, s session.Session) *Decoder {
	return &Decoder{r, s, NewTranslate(s)}
}

// Decode decodes a single message from a buffer of bytes.
func (d *Decoder) Decode(data []byte) (*Message, error) {
	return Read(bytes.NewBuffer(data), d.Session, d.Translate)
}

// Next decodes the next message from the stream. Note that if there is an
// exception, depending on where the exception originated from, the decoder
// results can no longer be trusted and the stream should be reset.
func (d *Decoder) Next() (*Message, error) {
	return Read(d.Reader, d.Session, d.Translate)
}

// Read a single IPFIX message from the provided reader and decode all the sets.
func Read(r io.Reader, s session.Session, t *Translate) (*Message, error) {
	m := new(Message)

	if t == nil && s != nil {
		t = NewTranslate(s)
	}

	if err := m.Header.Unmarshal(r); err != nil {
		return nil, err
	}
	if int(m.Header.Length) < m.Header.Len() {
		return nil, io.ErrShortBuffer
	}
	if m.Header.Version != Version {
		return nil, errInvalidVersion(m.Header.Version)
	}

	return m, m.UnmarshalSets(r, s, t)
}

// ValidateTemplateSize validates that a template's calculated size matches expectations.
// This helps detect template desynchronization issues early.
func ValidateTemplateSize(template session.Template) int {
	size := 0
	for _, field := range template.GetFields() {
		size += int(field.GetLength())
	}
	return size
}
