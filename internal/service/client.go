// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Davincible/goinsta"
	"github.com/tcnksm/go-input"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/pkg/spinner"
)

func makeClient(ctx context.Context, cfg config.Config, cfgPath string) (*goinsta.Instagram, error) {
	var cl *goinsta.Instagram

	uname, err := username()
	if err != nil {
		return nil, fmt.Errorf("username: %w", err)
	}

	sessFile := filepath.Join(cfgPath, fmt.Sprintf("%s.sess", uname))

	stop := spinner.Set("Trying to import previous session..", "", "yellow")

	cl, err = goinsta.Import(sessFile)

	stop()

	if err == nil {
		log.WithField(ctx, "session_file", sessFile).Info("Session imported")

		return cl, nil
	}

	pwd, err := password()
	if err != nil {
		return nil, fmt.Errorf("password: %w", err)
	}

	cl, err = login(uname, pwd)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	if cfg.StoreSession() {
		if err = cl.Export(sessFile); err != nil {
			log.WithError(ctx, err).Error("Failed to save session")
		}
	}

	return cl, nil
}

func login(uname, pwd string) (*goinsta.Instagram, error) {
	cl := goinsta.New(uname, pwd)

	stop := spinner.Set("Sending log in request..", "", "yellow")

	err := cl.Login()

	stop()

	switch {
	case errors.Is(err, nil):
		return cl, nil
	case errors.Is(err, goinsta.ErrChallengeRequired):
		var chErr *goinsta.ChallengeError

		if !errors.As(err, &chErr) {
			return nil, fmt.Errorf("failed to get challenge details: %w", err)
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

		stop = spinner.Set("Sending 2fa code..", "", "yellow")
		defer stop()

		if err = cl.TwoFactorInfo.Login2FA(code); err != nil {
			return nil, fmt.Errorf("login 2fa: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected: %w", err)
	}

	return cl, nil
}

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

func getPrompt(ask, key string) (string, error) {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	in, err := ui.Ask(ask,
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

	return in, nil
}

func challenge(cl *goinsta.Instagram, chURL string) (*goinsta.Instagram, error) {
	if err := cl.Challenge.ProcessOld(chURL); err != nil {
		return nil, fmt.Errorf("process challenge: %w", err)
	}

	ask := "What is SMS code for instagram?"
	key := "SMS code"

	code, err := getPrompt(ask, key)
	if err != nil {
		return nil, fmt.Errorf("get prompt: %w", err)
	}

	stop := spinner.Set("Sending security code..", "", "yellow")
	defer stop()

	if err = cl.Challenge.SendSecurityCode(code); err != nil {
		return nil, fmt.Errorf("send security code: %w", err)
	}

	return cl, nil
}
