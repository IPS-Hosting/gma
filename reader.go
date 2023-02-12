package gma

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type Reader struct {
	src Source
}

// Descriptions of newer addons are in JSON format
type jsonDescription struct {
	Description string
	Type        AddonType
	Tags        []AddonTag
}

func (r *Reader) readByte() (byte, error) {
	b, err := r.readBytes(1)
	return b[0], err
}

func (r *Reader) readBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := r.src.Read(b)
	return b, err
}

func (r *Reader) readUint32() (uint32, error) {
	b, err := r.readBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *Reader) readUint64() (uint64, error) {
	b, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (r *Reader) readInt32() (int32, error) {
	b, err := r.readBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(b)), nil
}

func (r *Reader) readString() (string, error) {
	var (
		err error
		b   = make([]byte, 1)
		str string
	)

	// Read byte by byte until we get a 0 or EOF
	for {
		_, err = r.src.Read(b)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if b[0] == 0 {
			break
		}
		str += string(b)
	}

	return str, err
}

// ReadAddon tries to read an addon from the source.
func (r *Reader) ReadAddon() (*Addon, error) {
	var (
		err               error
		ident             []byte
		parsedDescription *jsonDescription
		addon             = &Addon{}
	)

	_, err = r.src.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Ident
	ident, err = r.readBytes(4)
	if err != nil {
		return nil, err
	}
	if string(ident) != Ident {
		return nil, errors.New("not a valid GMA file")
	}

	// Format version
	addon.FormatVersion, err = r.readByte()
	if err != nil {
		return nil, err
	}
	if addon.FormatVersion > Version {
		return nil, errors.New("unsupported addon version")
	}

	// Steam ID
	addon.SteamID, err = r.readUint64()
	if err != nil {
		return nil, err
	}

	// Timestamp
	addon.Timestamp, err = r.readUint64()
	if err != nil {
		return nil, err
	}

	// Required content
	if addon.FormatVersion > 1 {
		var s string
		for {
			s, err = r.readString()
			if err != nil {
				return nil, err
			}
			if len(s) == 0 {
				break
			}
			addon.RequiredContent += s
		}
	}

	// Name
	addon.Name, err = r.readString()
	if err != nil {
		return nil, err
	}

	// Description
	addon.Description, err = r.readString()
	if err != nil {
		return nil, err
	}

	// Try to parse description as json. Ignore errors, as old addons have a plain text description
	parsedDescription = &jsonDescription{}
	err = json.Unmarshal([]byte(addon.Description), parsedDescription)
	if err == nil {
		addon.Description = parsedDescription.Description
		addon.Type = parsedDescription.Type
		addon.Tags = parsedDescription.Tags
	}

	// Author
	addon.Author, err = r.readString()
	if err != nil {
		return nil, err
	}

	// Version
	addon.Version, err = r.readInt32()
	if err != nil {
		return nil, err
	}

	// File entries
	var offset uint64 = 0
	for {
		entry := AddonFileEntry{}

		entry.ID, err = r.readUint32()
		if err != nil {
			return nil, err
		}
		if entry.ID == 0 {
			break
		}

		entry.Name, err = r.readString()
		if err != nil {
			return nil, err
		}

		entry.Size, err = r.readUint64()
		if err != nil {
			return nil, err
		}

		entry.CRC, err = r.readUint32()
		if err != nil {
			return nil, err
		}

		entry.Offset = offset
		offset += entry.Size

		addon.Files = append(addon.Files, entry)
	}

	addon.Src = r.src
	addon.FileBlockOffset, err = r.src.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	return addon, nil
}

func NewReader(src Source) *Reader {
	return &Reader{src: src}
}
