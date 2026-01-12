package utils

import (
	"sync"

	"simdokpol/web"
)

var (
	logoOnce  sync.Once
	logoBytes []byte
	logoErr   error
)

func loadLogoBytes() ([]byte, error) {
	logoOnce.Do(func() {
		logoBytes, logoErr = web.Assets.ReadFile("static/img/logo.png")
	})
	return logoBytes, logoErr
}
