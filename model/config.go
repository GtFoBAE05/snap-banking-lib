package model

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"snap-banking-lib/internal/crypto"
)

type Config struct {
	Banks map[BankCode]*BankConfig
}

type BankConfig struct {
	APIBaseURL     string
	PartnerBaseURL string
	Endpoints      map[EndpointKey]Endpoint
	APIKey         string
	APISecret      string
	ClientID       string
	ClientSecret   string
	PrivateKeyPEM  []byte
	PublicKeyPEM   []byte
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
}

func (c *Config) Validate() error {
	if len(c.Banks) == 0 {
		return errors.New("at least one bank must be configured")
	}

	for bankCode, bankConfig := range c.Banks {
		if err := bankConfig.Validate(bankCode); err != nil {
			return fmt.Errorf("invalid config for bank %s: %w", bankCode, err)
		}
	}

	return nil
}

func (bc *BankConfig) Validate(bank BankCode) error {
	if bc.APIBaseURL == "" {
		return errors.New("APIBaseURL is required")
	}

	bc.setDefaultEndpoints(bank)

	if len(bc.PrivateKeyPEM) == 0 {
		return errors.New("private key is required")
	}

	privateKey, err := crypto.ParsePrivateKey(bc.PrivateKeyPEM)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}
	bc.privateKey = privateKey

	if len(bc.PublicKeyPEM) > 0 {
		publicKey, err := crypto.ParsePublicKey(bc.PublicKeyPEM)
		if err != nil {
			return fmt.Errorf("invalid public key: %w", err)
		}
		bc.publicKey = publicKey
	}

	return nil
}

func (bc *BankConfig) setDefaultEndpoints(bank BankCode) {
	var defaults map[EndpointKey]Endpoint

	switch bank {
	case BankBCA:
		defaults = DefaultBCAEndpoints
	case BankBRI:
		defaults = DefaultBRIEndpoints
	default:
		return
	}

	if bc.Endpoints == nil {
		bc.Endpoints = defaults
		return
	}

	for k, v := range defaults {
		if _, ok := bc.Endpoints[k]; !ok {
			bc.Endpoints[k] = v
		}
	}
}

func (bc *BankConfig) PrivateKey() *rsa.PrivateKey {
	return bc.privateKey
}

func (bc *BankConfig) PublicKey() *rsa.PublicKey {
	return bc.publicKey
}

func (c *Config) GetBankConfig(bank BankCode) (*BankConfig, bool) {
	config, ok := c.Banks[bank]
	return config, ok
}

func (c *Config) HasBank(bank BankCode) bool {
	_, ok := c.Banks[bank]
	return ok
}
