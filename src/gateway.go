package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

// Gateway Declaration
type Gateway struct {
	Cache      *Cache
	Config     *Config
	AuthServer *server.Server
}

// Start App
// TO-DO: Recieve config on init
// TO-DO: Backend OAuth tokens to persistent and shared store
// TO-DO: Backend config to DB
func (g *Gateway) Start(config *Config) {
	g.Config = config

	manager := manage.NewDefaultManager()

	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte("keymaker"), jwt.SigningMethodHS512))

	clientStore := store.NewClientStore()

	clientStore.Set("222222", &models.Client{
		ID:     "222222",
		Secret: "22222222",
		Domain: "http://localhost:9000",
	})

	manager.MapClientStorage(clientStore)

	g.AuthServer = server.NewDefaultServer(manager)

	g.AuthServer.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Error: OAuth Server Internal", err.Error())
		return
	})

	g.AuthServer.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Error: OAuth Server Response", re.Error.Error())
	})

	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		err := g.AuthServer.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Access Token Request", r.Host, r.RequestURI)

		err := g.AuthServer.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/", g.handleProxy)

	g.Cache = InitializeCache()

	log.Println("Starting Periscope Gateway Port", config.Port)

	log.Fatal(http.ListenAndServe(config.Port, nil))
}

func (g *Gateway) handleProxy(w http.ResponseWriter, r *http.Request) {
	_, err := g.AuthServer.ValidationBearerToken(r)
	if err != nil {
		g.respondWithError(w, http.StatusForbidden, "Invalid Access Token")
		return
	}

	// PREFLOW

	// Get Proxy
	requestPath := r.URL.Path

	proxy, err := FindProxy(g.Config.Proxies, requestPath)
	if proxy == nil || err != nil {
		g.respondWithError(w, http.StatusNotFound, "Matching Proxy Not Found")
		return
	}

	log.Println("Request", r.Host, proxy.Method, proxy.Path)

	// Check Cache
	if len(proxy.Cache) > 0 {
		_, err = time.ParseDuration(proxy.Cache)
		if err == nil {
			proxyResponse := g.Cache.Get(requestPath)

			if proxyResponse.Content != nil {
				log.Println("Cache Hit", r.RequestURI)
				g.respondWithProxyResponse(w, proxyResponse)
				return
			}
		}
	}

	// REQUEST
	proxyURL, err := url.Parse(proxy.Target)
	if err != nil {
		g.respondWithError(w, http.StatusNotFound, "Invalid Proxy Target Format")
	}

	r.Host = proxyURL.Host
	r.URL.Host = proxyURL.Host
	r.URL.Scheme = proxyURL.Scheme
	r.RequestURI = ""

	s, _, _ := net.SplitHostPort(r.RemoteAddr)
	r.Header.Set("X-Forwarded-For", s)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		g.respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		g.respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	proxyResponse := ProxyResponse{
		Content:    content,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
	}

	log.Println("Response", requestPath, proxyResponse.StatusCode)

	g.respondWithProxyResponse(w, &proxyResponse)

	// POSTFLOW

	// Set Cache
	if len(proxy.Cache) > 0 {
		d, err := time.ParseDuration(proxy.Cache)
		if err == nil {
			g.Cache.Set(requestPath, proxyResponse, d)
			log.Println("Cached", r.RequestURI, proxy.Cache)
		}
	}
}

func (g *Gateway) respondWithProxyResponse(w http.ResponseWriter, proxyResponse *ProxyResponse) {
	for key, values := range proxyResponse.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	w.WriteHeader(proxyResponse.StatusCode)

	w.Write(proxyResponse.Content)
}

func (g *Gateway) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (g *Gateway) respondWithError(w http.ResponseWriter, code int, message string) {
	g.respondWithJSON(w, code, map[string]string{"error": message})

	log.Println("Error:", message)
}
