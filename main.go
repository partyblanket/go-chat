package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/partyblanket/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates",
			t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	envPort, exists := os.LookupEnv("PORT")
	if exists {
		fmt.Println("envirionment port: ", envPort)
		*addr = ":" + envPort
	}
	var tracerOn = flag.Bool("trace", false, "With tracing.")
	flag.Parse() // parse the flags
	r := newRoom()
	if *tracerOn {
		r.tracer = trace.New(os.Stdout)
	}
	// setup gomniauth
	// ID:

	gomniauth.SetSecurityKey("the_mountains_are_really_beautiful_today")
	gomniauth.WithProviders(
		facebook.New("key", "secret",
			"http://me.me.com:8080/auth/callback/facebook"),
		github.New("key", "secret",
			"http://localhost:8080/auth/callback/github"),
		google.New("452843566981-q8jut402jb2ncj2k9kvaih2n2ot7fpp8.apps.googleusercontent.com", "Wpug48rV1t6cOJki7-cwvHwl",
			"http://me.me.com:8080/auth/callback/google"),
	)

	http.Handle("/", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.Handle("/room", r)
	http.HandleFunc("/auth/", loginHandler)
	go r.run()
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
