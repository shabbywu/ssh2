package tempfile

import (
	"os"
	"path"
)

type TempFileManger struct {
	dir      string
	cleaners []func()
}

var cache = map[string]*TempFileManger{}

func GetManager(dir string) *TempFileManger {
	if m, ok := cache[dir]; ok {
		return m
	} else {
		cache[dir] = NewManager(dir)
		return cache[dir]
	}
}

func NewManager(dir string) *TempFileManger {
	if dir == "" {
		dir = os.TempDir()
	}
	return &TempFileManger{
		dir: dir,
	}
}

func (m *TempFileManger) TempFile(pattern string) (f *os.File, err error) {
	file, err := os.CreateTemp(m.dir, pattern)
	os.Chown(path.Join(m.dir, file.Name()), os.Getgid(), os.Getuid())
	m.cleaners = append(m.cleaners, func() {
		os.Remove(file.Name())
	})
	return file, err
}

func (m *TempFileManger) Clean() {
	for _, c := range m.cleaners {
		c()
	}
}
