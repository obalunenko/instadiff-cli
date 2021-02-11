// Package service implements instagram account operations and business logic.
package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheForgotten69/goinsta/v2"
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
		var chErr *goinsta.ChallengeError

		if !errors.As(err, &chErr) {
			return nil, fmt.Errorf("failed to login: %w", err)
		}

		cl, err = challenge(cl, chErr.Challenge.APIPath)
		if err != nil {
			return nil, fmt.Errorf("challenge: %w", err)
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
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	name, err := ui.Ask("What is your username?",
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
		return "", fmt.Errorf("username input: %w", err)
	}

	return name, nil
}

func password() (string, error) {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	pwd, err := ui.Ask("What is your password?",
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
		return "", fmt.Errorf("password input: %w", err)
	}

	return pwd, nil
}

func challenge(cl *goinsta.Instagram, chURL string) (*goinsta.Instagram, error) {
	if err := cl.Challenge.Process(chURL); err != nil {
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
