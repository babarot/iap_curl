package iap

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	envCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	envClientID    = "IAP_CLIENT_ID"
	envCurlCommand = "IAP_CURL_BIN"
)

// Config represents
type Config struct {
	Services []Service `json:"services"`

	path string
}

// Service is the URL and its Env pair
type Service struct {
	URL string `json:"url"`
	Env Env    `json:"env"`
}

// Env represents the environment variables needed to request to IAP-protected app
type Env struct {
	Credentials string `json:"GOOGLE_APPLICATION_CREDENTIALS"`
	ClientID    string `json:"IAP_CLIENT_ID"`
	Binary      string `json:"IAP_CURL_BIN"`
}

func configDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	default:
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	case "windows":
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data")
		}
	}
	dir = filepath.Join(dir, "iap_curl")

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return dir, fmt.Errorf("cannot create directory: %v", err)
	}

	return dir, nil
}

// Create creates config file if it doesn't exist
func (cfg *Config) Create() error {
	_, err := os.Stat(cfg.path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	f, err := os.Create(cfg.path)
	if err != nil {
		return err
	}

	// Insert sample config map as a default
	if len(cfg.Services) == 0 {
		cfg.Services = []Service{Service{
			URL: "https://iap-protected-app-url",
			Env: Env{
				Credentials: "/path/to/google-credentials.json",
				ClientID:    "foobar.apps.googleusercontent.com",
				Binary:      "curl",
			},
		}}
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

// Load loads config file to struct
func (cfg *Config) Load() error {
	dir, _ := configDir()
	file := filepath.Join(dir, "config.json")
	cfg.path = file
	_, err := os.Stat(cfg.path)
	if err != nil {
		return err
	}
	raw, _ := ioutil.ReadFile(cfg.path)
	return json.Unmarshal(raw, cfg)
}

func (cfg *Config) getEnvFromFile(url string) (env Env, err error) {
	u1, _ := neturl.Parse(url)
	for _, service := range cfg.Services {
		u2, _ := neturl.Parse(service.URL)
		if u1.Host == u2.Host {
			return service.Env, nil
		}
	}
	err = fmt.Errorf("%s: no such host in config file", u1.Host)
	return
}

// GetEnv returns Env includes url
func (cfg *Config) GetEnv(url string) (env Env, err error) {
	env, _ = cfg.getEnvFromFile(url)
	credentials := os.Getenv(envCredentials)
	clientID := os.Getenv(envClientID)
	binary := os.Getenv(envCurlCommand)
	if credentials == "" {
		credentials, _ = homedir.Expand(env.Credentials)
	}
	if clientID == "" {
		clientID = env.ClientID
	}
	if binary == "" {
		binary = env.Binary
	}
	if credentials == "" {
		return env, fmt.Errorf("%s is missing", envCredentials)
	}
	if clientID == "" {
		return env, fmt.Errorf("%s is missing", envClientID)
	}
	if binary == "" {
		binary = "curl"
	}
	return Env{
		Credentials: credentials,
		ClientID:    clientID,
		Binary:      binary,
	}, nil
}

// GetURLs returns URLs described in config file
func (cfg *Config) GetURLs() (list []string) {
	for _, service := range cfg.Services {
		list = append(list, service.URL)
	}
	return
}

// Edit edits config file
// If it doesn't exist, it will be automatically created
func (cfg *Config) Edit() error {
	if err := cfg.Create(); err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	command := fmt.Sprintf("%s %s", editor, cfg.path)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Registered returns true if url exists in config file
func (cfg *Config) Registered(url string) bool {
	u1, _ := neturl.Parse(url)
	for _, service := range cfg.Services {
		u2, _ := neturl.Parse(service.URL)
		if u1.Host == u2.Host {
			return true
		}
	}
	return false
}

// Register registers service to config file
func (cfg *Config) Register(s Service) error {
	cfg.Services = append(cfg.Services, s)
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return err
	}
	file, err := os.OpenFile(cfg.path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)
	w.Write(out.Bytes())
	return w.Flush()
}
