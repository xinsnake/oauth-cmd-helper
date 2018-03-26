package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/skratchdot/open-golang/open"
)

type token struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

var port = fmt.Sprintf(":%s", os.Getenv("PORT"))
var callbackURI = fmt.Sprintf("http://localhost%s/callback", port)
var postURI = fmt.Sprintf("%s/oauth2/token", os.Getenv("BASE_URI"))

func waitForEnter() {
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func handler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Fatal("Unable to get code, please try again")
	}
	postBody := bytes.NewBufferString(url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {os.Getenv("CLIENT_ID")},
		"scope":        {os.Getenv("SCOPE")},
		"redirect_uri": {callbackURI},
		"code":         {code},
	}.Encode())
	postReq, err := http.NewRequest(http.MethodPost, postURI, postBody)
	if err != nil {
		log.Fatalf("Unable to construct request: %v", err)
	}
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if os.Getenv("CLIENT_SECRET") != "" {
		authString := base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))))
		postReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", authString))
	}
	client := &http.Client{}
	postResp, err := client.Do(postReq)
	if err != nil {
		log.Fatalf("Unable to get token: %v", err)
	} else if postResp.StatusCode != 200 {
		log.Fatalf("Unable to get token: %v", postResp)
	}
	postRespBody, err := ioutil.ReadAll(postResp.Body)
	if err != nil {
		log.Fatalf("Error reading body: %v", err)
	}

	var t token
	json.Unmarshal([]byte(postRespBody), &t)
	output, _ := json.MarshalIndent(t, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(output))
}

func main() {
	log.Printf("Please set the call back URI to:\n==> %s\nbefore pressing Enter to continue", callbackURI)
	waitForEnter()

	openParams := url.Values{
		"response_type": {"code"},
		"client_id":     {os.Getenv("CLIENT_ID")},
		"scope":         {os.Getenv("SCOPE")},
		"redirect_uri":  {callbackURI},
	}.Encode()
	var openURI = fmt.Sprintf("%s/oauth2/authorize?%s", os.Getenv("BASE_URI"), openParams)
	log.Printf("Now your browser should open with URI:\n==> %s\nif not please open it manually", openURI)
	open.Run(openURI)

	http.HandleFunc("/callback", handler)
	log.Fatal(http.ListenAndServe(port, nil))
}
