package model

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func LoadFromEnv(dotenvFiles ...string) (*Config, error) {
	if len(dotenvFiles) > 0 {
		if err := godotenv.Load(dotenvFiles...); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading dotenv: %w", err)
		}
	}

	banksRaw := os.Getenv("SNAP_BANKS")
	if banksRaw == "" {
		return nil, fmt.Errorf("SNAP_BANKS is required (e.g. SNAP_BANKS=BCA,BRI)")
	}

	banks := make(map[BankCode]*BankConfig)
	for _, raw := range splitComma(banksRaw) {
		code := BankCode(strings.ToUpper(raw))
		bc, err := bankFromEnv(strings.ToUpper(raw))
		if err != nil {
			return nil, fmt.Errorf("bank %s: %w", code, err)
		}
		banks[code] = bc
	}

	return &Config{Banks: banks}, nil
}

func bankFromEnv(bank string) (*BankConfig, error) {
	p := "SNAP_" + bank + "_"
	get := func(k string) string { return strings.TrimSpace(os.Getenv(p + k)) }

	privPEM, err := readPEMFile(get("PRIVATE_KEY_PATH"), bank+" private key")
	if err != nil {
		return nil, err
	}

	pubPEM, err := readPEMFile(get("PUBLIC_KEY_PATH"), bank+" public key")
	if err != nil {
		return nil, err
	}

	endpoints := endpointsFromEnv(p)

	return &BankConfig{
		APIBaseURL:    get("API_BASE_URL"),
		APIKey:        get("API_KEY"),
		APISecret:     get("API_SECRET"),
		ClientID:      get("CLIENT_ID"),
		ClientSecret:  get("CLIENT_SECRET"),
		PrivateKeyPEM: privPEM,
		PublicKeyPEM:  pubPEM,
		Endpoints:     endpoints,
	}, nil
}

func endpointsFromEnv(prefix string) map[EndpointKey]Endpoint {
	allKeys := []EndpointKey{
		EndpointAccessToken,
		EndpointBalanceInquiry,
		EndpointBankStatement,
		EndpointVirtualAccountInquiry,
		EndpointVirtualAccountInquiryStatus,
		EndpointVirtualAccountPayment,
		EndpointVirtualAccountIntrabankInquiry,
		EndpointVirtualAccountIntrabankPaymentNotification,
		EndpointVirtualAccountIntrabankPayment,
		EndpointQRMPMGenerate,
		EndpointQRMPMInquiry,
		EndpointQRMPMRefund,
		EndpointQRISNotification,
		EndpointExternalAccountInquiry,
		EndpointInterbankTransfer,
		EndpointInternalAccountInquiry,
		EndpointIntrabankTransfer,
		EndpointTransactionStatusInquiry,
	}

	endpoints := make(map[EndpointKey]Endpoint)
	for _, key := range allKeys {
		envKey := strings.ToUpper(string(key))
		path := strings.TrimSpace(os.Getenv(prefix + "ENDPOINT_" + envKey + "_PATH"))
		if path != "" {
			endpoints[key] = Endpoint{Path: path}
		}
	}

	if len(endpoints) == 0 {
		return nil
	}
	return endpoints
}

func LoadFromYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading yaml config: %w", err)
	}

	var raw struct {
		HTTPTimeout string                    `yaml:"http_timeout"`
		Banks       map[string]yamlBankConfig `yaml:"banks"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing yaml config: %w", err)
	}

	banks := make(map[BankCode]*BankConfig)
	for rawCode, yb := range raw.Banks {
		code := BankCode(strings.ToUpper(rawCode))
		bc, err := yb.toBankConfig(strings.ToUpper(rawCode))
		if err != nil {
			return nil, fmt.Errorf("bank %s: %w", code, err)
		}
		banks[code] = bc
	}

	return &Config{Banks: banks}, nil
}

type yamlBankConfig struct {
	APIBaseURL     string            `yaml:"api_base_url"`
	APIKey         string            `yaml:"api_key"`
	APISecret      string            `yaml:"api_secret"`
	ClientID       string            `yaml:"client_id"`
	ClientSecret   string            `yaml:"client_secret"`
	PrivateKeyPath string            `yaml:"private_key_path"`
	PublicKeyPath  string            `yaml:"public_key_path"`
	Endpoints      map[string]string `yaml:"endpoints"`
}

func (yb yamlBankConfig) toBankConfig(bank string) (*BankConfig, error) {
	privPEM, err := readPEMFile(yb.PrivateKeyPath, bank+" private key")
	if err != nil {
		return nil, err
	}

	pubPEM, err := readPEMFile(yb.PublicKeyPath, bank+" public key")
	if err != nil {
		return nil, err
	}

	var endpoints map[EndpointKey]Endpoint
	if len(yb.Endpoints) > 0 {
		endpoints = make(map[EndpointKey]Endpoint, len(yb.Endpoints))
		for k, path := range yb.Endpoints {
			endpoints[EndpointKey(k)] = Endpoint{Path: path}
		}
	}

	return &BankConfig{
		APIBaseURL:    yb.APIBaseURL,
		APIKey:        yb.APIKey,
		APISecret:     yb.APISecret,
		ClientID:      yb.ClientID,
		ClientSecret:  yb.ClientSecret,
		PrivateKeyPEM: privPEM,
		PublicKeyPEM:  pubPEM,
		Endpoints:     endpoints,
	}, nil
}

func readPEMFile(path, label string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	pem, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot read %q: %w", label, path, err)
	}
	return pem, nil
}

func parseDuration(s string, defaultVal time.Duration) (time.Duration, error) {
	if s == "" {
		return defaultVal, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q (e.g. \"30s\"): %w", s, err)
	}
	return d, nil
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
