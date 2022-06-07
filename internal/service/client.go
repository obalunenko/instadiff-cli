// Package service implements instagram account operations and business logic.
package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Davincible/goinsta"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"

	"github.com/obalunenko/instadiff-cli/internal/config"
)

func makeClient(cfg config.Config, cfgPath string) (*goinsta.Instagram, error) {
	var cl *goinsta.Instagram

	uname, err := username()
	if err != nil {
		return nil, fmt.Errorf("username: %w", err)
	}

	sessFile := filepath.Join(cfgPath, fmt.Sprintf("%s.sess", uname))

	if cl, err = goinsta.Import(sessFile); err == nil {
		log.Infof("session imported from file: %s", sessFile)

		return cl, nil
	}

	pwd, err := password()
	if err != nil {
		return nil, fmt.Errorf("password: %w", err)
	}

	cl = goinsta.New(uname, pwd)

	if err = cl.Login(); err != nil {
		switch {
		case errors.Is(err, goinsta.ErrChallengeRequired):
			var chErr *goinsta.ChallengeError

			if !errors.As(err, &chErr) {
				return nil, fmt.Errorf("failed to login: %w", err)
			}

			cl, err = challenge(cl, chErr.Challenge.APIPath)
			if err != nil {
				return nil, fmt.Errorf("challenge: %w", err)
			}

		case errors.Is(err, goinsta.Err2FARequired) || errors.Is(err, goinsta.Err2FANoCode):
			var code string

			code, err = twoFactorCode()
			if err != nil {
				return nil, fmt.Errorf("2fa ocde: %w", err)
			}

			if err = cl.TwoFactorInfo.Login2FA(code); err != nil {
				return nil, fmt.Errorf("login 2fa: %w", err)
			}
		}
	}

	if cfg.StoreSession() {
		if err = cl.Export(sessFile); err != nil {
			log.Errorf("save session: %v", err)
		}
	}

	return cl, nil
}

// ErrEmptyInput returned in case when user input is empty.
var ErrEmptyInput = errors.New("should not be empty")

func username() (string, error) {
	ask := "What is your username?"
	key := "username"

	return getPrompt(ask, key)
}

func password() (string, error) {
	ask := "What is your password?"
	key := "password"

	return getPrompt(ask, key)
}

func twoFactorCode() (string, error) {
	ask := "What is your two factor code?"
	key := "2fa code"

	return getPrompt(ask, key)
}

func getPrompt(ask string, key string) (string, error) {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	input, err := ui.Ask(ask,
		&input.Options{
			Default:     "",
			Loop:        true,
			Required:    true,
			HideDefault: false,
			HideOrder:   false,
			Hide:        false,
			Mask:        false,
			MaskDefault: false,
			MaskVal:     "",
			ValidateFunc: func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return ErrEmptyInput
				}

				return nil
			},
		})
	if err != nil {
		return "", fmt.Errorf("%s input: %w", key, err)
	}

	return input, nil
}

func challenge(cl *goinsta.Instagram, chURL string) (*goinsta.Instagram, error) {
	if err := cl.Challenge.ProcessOld(chURL); err != nil {
		return nil, fmt.Errorf("process challenge: %w", err)
	}

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	code, err := ui.Ask("What is SMS code for instagram?",
		&input.Options{
			Default:     "",
			Loop:        true,
			Required:    true,
			HideDefault: false,
			HideOrder:   false,
			Hide:        false,
			Mask:        false,
			MaskDefault: false,
			MaskVal:     "",
			ValidateFunc: func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return ErrEmptyInput
				}

				return nil
			},
		})
	if err != nil {
		return nil, fmt.Errorf("process input: %w", err)
	}

	if err = cl.Challenge.SendSecurityCode(code); err != nil {
		return nil, fmt.Errorf("send security code: %w", err)
	}

	return cl, nil
}
