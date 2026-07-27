package sqldump

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/klauspost/pgzip"
)

// OpenReader abre path para leitura em streaming, descomprimindo com pgzip
// (paralelo, multi-core) se o nome terminar em .gz.
func OpenReader(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ".gz") {
		gz, err := pgzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &gzReadCloser{gz: gz, f: f}, nil
	}
	return f, nil
}

// OpenWriter abre path para escrita em streaming, comprimindo com pgzip
// (paralelo, multi-core) se o nome terminar em .gz.
func OpenWriter(path string) (io.WriteCloser, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ".gz") {
		gz := pgzip.NewWriter(f)
		return &gzWriteCloser{gz: gz, f: f}, nil
	}
	return f, nil
}

type gzReadCloser struct {
	gz *pgzip.Reader
	f  *os.File
}

func (g *gzReadCloser) Read(p []byte) (int, error) { return g.gz.Read(p) }

func (g *gzReadCloser) Close() error {
	g.gz.Close()
	return g.f.Close()
}

type gzWriteCloser struct {
	gz *pgzip.Writer
	f  *os.File
}

func (g *gzWriteCloser) Write(p []byte) (int, error) { return g.gz.Write(p) }

func (g *gzWriteCloser) Close() error {
	if err := g.gz.Close(); err != nil {
		g.f.Close()
		return err
	}
	return g.f.Close()
}

// LineScanner itera um io.Reader linha a linha, preservando o terminador \n
// (equivalente a iterar um arquivo texto aberto em Python, inclusive a
// última linha sem \n final).
type LineScanner struct {
	r   *bufio.Reader
	err error
}

func NewLineScanner(r io.Reader) *LineScanner {
	return &LineScanner{r: bufio.NewReaderSize(r, 1<<20)}
}

// Next devolve a próxima linha e true, ou "", false quando o arquivo acaba.
func (s *LineScanner) Next() (string, bool) {
	if s.err != nil {
		return "", false
	}
	line, err := s.r.ReadString('\n')
	if len(line) == 0 && err != nil {
		s.err = err
		return "", false
	}
	if err != nil && err != io.EOF {
		s.err = err
		return "", false
	}
	return line, true
}

// Err devolve o primeiro erro de leitura encontrado (io.EOF não conta como erro).
func (s *LineScanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
