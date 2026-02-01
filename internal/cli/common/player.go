package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

// PlayFile plays an audio file through the system speakers.
// Supported formats: mp3, wav
func PlayFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	switch ext {
	case ".mp3":
		return playMP3(file)
	case ".wav":
		return playWAV(file)
	default:
		return fmt.Errorf("unsupported audio format: %s", ext)
	}
}

func playMP3(r io.Reader) error {
	decoder, err := mp3.NewDecoder(r)
	if err != nil {
		return fmt.Errorf("cannot decode mp3: %w", err)
	}

	otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   decoder.SampleRate(),
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return fmt.Errorf("cannot create audio context: %w", err)
	}
	<-ready

	player := otoCtx.NewPlayer(decoder)
	defer player.Close()

	player.Play()
	for player.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// wavHeader represents a minimal WAV file header
type wavHeader struct {
	SampleRate   uint32
	NumChannels  uint16
	BitsPerSample uint16
	DataSize     uint32
}

func parseWAVHeader(r io.Reader) (*wavHeader, error) {
	// Read RIFF header
	var riffHeader [12]byte
	if _, err := io.ReadFull(r, riffHeader[:]); err != nil {
		return nil, fmt.Errorf("cannot read RIFF header: %w", err)
	}
	if string(riffHeader[0:4]) != "RIFF" || string(riffHeader[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	header := &wavHeader{}

	// Read chunks until we find 'data'
	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(r, chunkHeader[:]); err != nil {
			return nil, fmt.Errorf("cannot read chunk header: %w", err)
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		if chunkID == "fmt " {
			// Read fmt chunk
			fmtData := make([]byte, chunkSize)
			if _, err := io.ReadFull(r, fmtData); err != nil {
				return nil, fmt.Errorf("cannot read fmt chunk: %w", err)
			}
			// audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
			header.NumChannels = binary.LittleEndian.Uint16(fmtData[2:4])
			header.SampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			// byteRate := binary.LittleEndian.Uint32(fmtData[8:12])
			// blockAlign := binary.LittleEndian.Uint16(fmtData[12:14])
			header.BitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])
		} else if chunkID == "data" {
			header.DataSize = chunkSize
			break
		} else {
			// Skip unknown chunk
			if _, err := io.CopyN(io.Discard, r, int64(chunkSize)); err != nil {
				return nil, fmt.Errorf("cannot skip chunk: %w", err)
			}
		}
	}

	return header, nil
}

func playWAV(r io.Reader) error {
	header, err := parseWAVHeader(r)
	if err != nil {
		return err
	}

	var format oto.Format
	switch header.BitsPerSample {
	case 8:
		format = oto.FormatUnsignedInt8
	case 16:
		format = oto.FormatSignedInt16LE
	case 32:
		format = oto.FormatFloat32LE
	default:
		return fmt.Errorf("unsupported bits per sample: %d", header.BitsPerSample)
	}

	otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   int(header.SampleRate),
		ChannelCount: int(header.NumChannels),
		Format:       format,
	})
	if err != nil {
		return fmt.Errorf("cannot create audio context: %w", err)
	}
	<-ready

	// Limit reader to data size
	dataReader := io.LimitReader(r, int64(header.DataSize))

	player := otoCtx.NewPlayer(dataReader)
	defer player.Close()

	player.Play()
	for player.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}
