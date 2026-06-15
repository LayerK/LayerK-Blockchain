package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
)

func CompressGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	if _, err := gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return buffer.Bytes(), nil
}

func DecompressGzip(data []byte) ([]byte, error) {
	return decompressGzip(data, 0)
}

func DecompressGzipWithLimit(data []byte, maxDecompressedBytes int64) ([]byte, error) {
	if maxDecompressedBytes < 0 {
		return nil, fmt.Errorf("invalid gzip decompression limit: %d", maxDecompressedBytes)
	}
	return decompressGzip(data, maxDecompressedBytes)
}

func decompressGzip(data []byte, maxDecompressedBytes int64) ([]byte, error) {
	buffer := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	var reader io.Reader = gzipReader
	if maxDecompressedBytes > 0 {
		limit := maxDecompressedBytes
		if limit < math.MaxInt64 {
			limit++
		}
		reader = io.LimitReader(gzipReader, limit)
	}
	decompressData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}
	if maxDecompressedBytes > 0 && int64(len(decompressData)) > maxDecompressedBytes {
		return nil, fmt.Errorf("decompressed data exceeds limit of %d bytes", maxDecompressedBytes)
	}
	return decompressData, nil
}
