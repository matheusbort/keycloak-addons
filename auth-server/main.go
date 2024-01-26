package main

import (
	"bytes"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

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

	r.POST("/otp/request/:phone", func(c *gin.Context) {
		session := sessions.Default(c)
		rand.Seed(time.Now().UnixNano())
		randomSixDigits := rand.Intn(900000) + 100000
		phone := c.Param("phone")
		otp := strconv.Itoa(randomSixDigits)
		session.Set("otp", otp)
		_ = session.Save()

		sendSMS(phone, otp)
		c.JSON(200, "sms sent")
	})

	r.POST("/otp/verify/:otp", func(c *gin.Context) {
		tokenOpts := gocloak.TokenOptions{
			ClientID:           gocloak.StringP(os.Getenv("CLIENT_ID")),
			ClientSecret:       gocloak.StringP(os.Getenv("CLIENT_SECRET")),
			GrantType:          gocloak.StringP("urn:ietf:params:oauth:grant-type:token-exchange"),
			RequestedTokenType: gocloak.StringP("urn:ietf:params:oauth:token-type:refresh_token"),
			RequestedSubject:   gocloak.StringP("math"),
			// Audience:           gocloak.StringP("public"), //
		}

		session := sessions.Default(c)
		sessionOtp := session.Get("otp")
		otp := c.Param("otp")

		if sessionOtp != otp {
			c.JSON(401, fmt.Sprintf("unmatch token"))
			return
		}

		token, err := kc.GetToken(c, os.Getenv("REALM_NAME"), tokenOpts)
		if err != nil {
			panic("Oh no!, failed to exchange token :(")
		}

		c.JSON(200, token)
	})

	r.POST("user/update/:id", func(c *gin.Context) {
		//{
		//    "id": "56a23d4e-2c32-4968-9834-c047ea6769b3",
		//    "username": "math2",
		//    "enabled": true,
		//    "firstName": "matheus",
		//    "lastName": "bortoletto",
		//    "email": "matheus1@suaquadra.com.br",
		//    "attributes": {
		//        "mobilePhone": [
		//            "+55199998430066"
		//        ],
		//        "mobilePhoneVerified": [
		//            "true"
		//        ]
		//    }
		//}

		var user gocloak.User
		_ = c.ShouldBindJSON(&user)
		token, _ := kc.LoginClient(c, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("REALM_NAME"))

		user.ID = gocloak.StringP(c.Param("id"))
		err := kc.UpdateUser(c, token.AccessToken, os.Getenv("REALM_NAME"), user)
		if err != nil {
			c.JSON(400, err)
			return
		}

		c.JSON(200, user)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func sendSMS(to string, token string) {
	// Set up the URL and authentication for the Twilio API
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", os.Getenv("TWILIO_PASSWD"))
	username := os.Getenv("TWILIO_USER")
	password := os.Getenv("TWILIO_PASSWD")

	// Create the data to be sent in the POST request
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", "+15108582511")
	data.Set("Body", "Seu token de acesso é: "+token)

	// Create a new HTTP client
	client := &http.Client{}

	// Create the POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		panic(err)
	}

	// Set the content type and basic authentication header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(username, password)

	// Perform the POST request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
