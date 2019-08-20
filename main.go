package main

import (
	"flag"
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
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	envPort, exists := os.LookupEnv("PORT")

	if exists {
		*addr = ":" + envPort
	}

	var tracerOn = flag.Bool("trace", false, "With tracing.")

	r := newRoom()
	if *tracerOn {
		r.tracer = trace.New(os.Stdout)
	}
	flag.Parse() // parse the flags
	// setup gomniauth
	gomniauth.SetSecurityKey("the_mountains_are_really_beautiful_today")
	gomniauth.WithProviders(
		facebook.New("key", "secret",
			"https://enigmatic-ravine-50615.herokuapp.com/auth/callback/facebook"),
		github.New("key", "secret",
			"https://enigmatic-ravine-50615.herokuapp.com/auth/callback/github"),
		google.New("452843566981-q8jut402jb2ncj2k9kvaih2n2ot7fpp8.apps.googleusercontent.com", "Wpug48rV1t6cOJki7-cwvHwl",
			"https://enigmatic-ravine-50615.herokuapp.com/auth/callback/google"),
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
