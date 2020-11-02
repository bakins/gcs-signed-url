package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"gopkg.in/alecthomas/kingpin.v2"

	"cloud.google.com/go/storage"
)

func main() {
	serviceAccountFile := kingpin.Flag("service-account-file", "path to service account file - default is $GOOGLE_APPLICATION_CREDENTIALS").
		Envar("GOOGLE_APPLICATION_CREDENTIALS").
		Required().
		ExistingFile()

	expires := kingpin.Flag("ttl", "url time to live").
		Envar("SIGNED_URL_TTL").
		Default("30m").
		Duration()

	method := kingpin.Flag("method", "HTTP method to sign").
		Envar("SIGNED_URL_METHOD").
		Default(http.MethodGet).
		String()

	object := kingpin.Arg("object", "gcs object url").
		Required().
		String()

	kingpin.Parse()

	log.SetFlags(0)

	if serviceAccountFile == nil || *serviceAccountFile == "" {
		log.Fatal("must provide service account file")
	}

	if !strings.HasPrefix(*object, "gs://") {
		*object = "gs://" + *object
	}
	u, err := url.Parse(*object)
	if err != nil {
		log.Fatal("failed to parse object url")
	}

	jsonKey, err := ioutil.ReadFile(*serviceAccountFile)
	if err != nil {
		log.Fatalf("failed to read service account file: %v", err)
	}

	jwtConf, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		log.Fatalf("failed to create jwt config from %s %v", *serviceAccountFile, err)
	}

	signedOptions := storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         *method,
		GoogleAccessID: jwtConf.Email,
		PrivateKey:     jwtConf.PrivateKey,
		Expires:        time.Now().Add(*expires),
	}

	signedURL, err := storage.SignedURL(u.Host, strings.TrimPrefix(u.Path, "/"), &signedOptions)
	if err != nil {
		log.Fatalf("failed to create signed url %s %v", u.String(), err)
	}

	fmt.Println(signedURL)
}
