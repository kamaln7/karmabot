package webui

import (
	"log"
	"time"

	"github.com/pquerna/otp/totp"

	//	"github.com/kamaln7/karmabot/database"
)

var (
	ll      *log.Logger
	totpKey string
)

// Init initiates the web ui config
func Init(logger *log.Logger, key, addr, path string) {
	ll = logger
	totpKey = key

	if totpKey == "" {
		newKey, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "karmabot",
			AccountName: "slack",
		})

		if err != nil {
			ll.Fatalf("an error occurred while generating a TOTP key: %s\n", err.Error())
		} else {
			ll.Fatalf("please use the following TOTP key (`karmabot -totp <key>`): %s\n", newKey.Secret())
		}
	}
}

// GetToken generates and returns a TOTP token
func GetToken() (string, error) {
	return totp.GenerateCode(totpKey, time.Now())
}

// Listen starts the web ui
func Listen() {
}
