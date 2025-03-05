package main

import (
	"log"
	"fmt"
	
	"github.com/Tootoohk/TwitterAPI/client"
	"github.com/Tootoohk/TwitterAPI/models"
	"github.com/Tootoohk/TwitterAPI/utils"
)

func main() {
	proxy := "user:pass@host:port"
	authToken := "auth_token_here"
	// Create account
	account := client.NewAccount(authToken, "", proxy)

	// Or with detailed logging
	verboseConfig := models.NewConfig()
	verboseConfig.LogLevel = utils.LogLevelDebug

	twitter, err := client.NewTwitter(account, verboseConfig)
	if err != nil {
		// Handle error your way
		log.Fatal(err)
	}

	info, status := twitter.IsValid()
	if status.Error != nil {
		log.Fatal(status.Error)
	}
	fmt.Println(info)
}
