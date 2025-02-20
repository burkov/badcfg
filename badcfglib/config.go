package badcfglib

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/tobischo/gokeepasslib/v3"
)

const tomlConfigFileName = "dev-config.toml"
const kdbxConfigFileName = "dev-config.kdbx"
const kdbxPasswordFileName = ".password-dev"

type BadConfig struct {
	TomlConfig map[string]string
	KdbxConfig map[string]string
}

type ConfigFileEntryType int

const (
	TomlConfig ConfigFileEntryType = iota
	KdbxConfig
)

type ConfigFileEntry struct {
	Key   string
	Value string
	Type  ConfigFileEntryType
}

func (c *BadConfig) Values() func(func(value ConfigFileEntry) bool) {
	return func(yield func(value ConfigFileEntry) bool) {
		keys := make([]string, 0, len(c.TomlConfig)+len(c.KdbxConfig))
		for k := range c.TomlConfig {
			keys = append(keys, k)
		}
		for k := range c.KdbxConfig {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			tomlValue, ok := c.TomlConfig[k]
			if ok {
				if !yield(ConfigFileEntry{Key: k, Value: tomlValue, Type: TomlConfig}) {
					return
				}
			} else {
				kdbxValue, ok := c.KdbxConfig[k]
				if ok {
					if !yield(ConfigFileEntry{Key: k, Value: kdbxValue, Type: KdbxConfig}) {
						return
					}
				}
			}
		}
	}
}

func ReadConfig() (*BadConfig, error) {
	locations, err := locateFiles()
	if err != nil {
		return nil, fmt.Errorf("error locating files: %v", err)
	}

	devConfig, err := readTomlConfig(locations.devConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading dev-config.toml: %v", err)
	}

	kdbxConfig, err := readKdbxConfig(locations.kdbxPath, locations.kdbxPasswordPath)
	if err != nil {
		return nil, fmt.Errorf("error reading kdbx file: %v", err)
	}

	return &BadConfig{
		TomlConfig: devConfig,
		KdbxConfig: kdbxConfig,
	}, nil
}

func readTomlConfig(path string) (map[string]string, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
		return nil, fmt.Errorf("error reading toml file: %v", err)
	}

	result := make(map[string]string)

	for _, key := range k.Keys() {
		result[key] = fmt.Sprintf("%v", k.Get(key))
	}

	return result, nil
}

type rootOnly struct {
	Root gokeepasslib.RootData `xml:"Root"`
}

func readKdbxConfig(dbPath string, passwordPath string) (map[string]string, error) {
	file, err := os.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening kdbx file: %v", err)
	}
	passwordBytes, err := os.ReadFile(passwordPath)
	if err != nil {
		return nil, fmt.Errorf("error reading password file: %v", err)
	}
	password := string(passwordBytes)
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	_ = gokeepasslib.NewDecoder(file).Decode(db)

	db.UnlockProtectedEntries()

	result := make(map[string]string)

	for _, group := range db.Content.Root.Groups {
		for _, entry := range group.Entries {
			result[entry.GetTitle()] = entry.GetPassword()
		}
	}

	return result, nil
}

type fileLocations struct {
	devConfigPath    string
	kdbxPath         string
	kdbxPasswordPath string
}

func locateFiles() (*fileLocations, error) {
	devConfigPath, err := findDevConfigPath()
	if err != nil {
		return nil, fmt.Errorf("error finding %v: %v", tomlConfigFileName, err)
	}
	kdbxPath, err := findSameDirFile(devConfigPath, kdbxConfigFileName)
	if err != nil {
		return nil, fmt.Errorf("error finding %v: %v", kdbxConfigFileName, err)
	}
	kdbxPasswordPath, err := findSameDirFile(devConfigPath, kdbxPasswordFileName)
	if err != nil {
		return nil, fmt.Errorf("error finding %v: %v", kdbxPasswordFileName, err)
	}
	return &fileLocations{
		devConfigPath:    devConfigPath,
		kdbxPath:         kdbxPath,
		kdbxPasswordPath: kdbxPasswordPath,
	}, nil
}

func findDevConfigPath() (string, error) {
	homedir, _ := os.UserHomeDir()
	var pathsForSearch = []string{
		"./" + tomlConfigFileName,
		"./config/" + tomlConfigFileName,
		homedir + "/IdeaProjects/JetProfile/config/" + tomlConfigFileName,
	}
	for _, location := range pathsForSearch {
		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}
	return "", fmt.Errorf("%v not found", tomlConfigFileName)
}

func findSameDirFile(path string, fileName string) (string, error) {
	dir := filepath.Dir(path)
	newPath := filepath.Join(dir, fileName)
	if _, err := os.Stat(newPath); err != nil {
		return "", fmt.Errorf("error finding file: %v", err)
	}
	return newPath, nil
}
