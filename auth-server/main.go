package main

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/joho/godotenv"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	godotenv.Load()

	kc := gocloak.NewClient(os.Getenv("KEYCLOAK_URL"))

	r.POST("/signup", func(c *gin.Context) {
		var user gocloak.User
		_ = c.ShouldBindJSON(&user)

		token, err := kc.LoginClient(c, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("REALM_NAME"))
		if err != nil {
			panic("Something wrong with the credentials or url")
		}

		_, err = kc.CreateUser(c, token.AccessToken, os.Getenv("REALM_NAME"), user)
		if err != nil {
			panic("Oh no!, failed to create user :(")
		}
	})

	r.POST("/login", func(c *gin.Context) {
		tokenOpts := gocloak.TokenOptions{
			ClientID:           gocloak.StringP(os.Getenv("CLIENT_ID")),
			ClientSecret:       gocloak.StringP(os.Getenv("CLIENT_SECRET")),
			GrantType:          gocloak.StringP("urn:ietf:params:oauth:grant-type:token-exchange"),
			RequestedTokenType: gocloak.StringP("urn:ietf:params:oauth:token-type:refresh_token"),
			RequestedSubject:   gocloak.StringP("math"),
			// Audience:           gocloak.StringP("public"), //
		}
		token, err := kc.GetToken(c, os.Getenv("REALM_NAME"), tokenOpts)
		if err != nil {
			panic("Oh no!, failed to exchange token :(")
		}

		c.JSON(200, token)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
