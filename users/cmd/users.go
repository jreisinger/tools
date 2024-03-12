package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/user"
	"strconv"
	"strings"
)

const (
	scheme  = "http://"
	baseURL = "/users/"
	maxUid  = 2000 // list only users with lower UID than this
)

func main() {
	server := flag.Bool("server", false, "run API server exposing system users")
	host := flag.String("host", "localhost", "network host")
	port := flag.String("port", "8080", "network port")
	flag.Parse()
	if *server {
		runServer(*host, *port)
	} else {
		runClient(*host, *port, flag.Args())
	}
}

func runClient(host, port string, usernames []string) {
	log.SetPrefix("users: ")
	log.SetFlags(0)

	addr := net.JoinHostPort(host, port)

	if len(usernames) == 0 {
		u := scheme + addr + baseURL
		b, err := getBody(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", b)
	}

	for _, name := range usernames {
		u := scheme + addr + baseURL + name
		b, err := getBody(u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", b)
	}
}

func getBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func runServer(host, port string) {
	http.HandleFunc("/users/", getUser)
	addr := net.JoinHostPort(host, port)
	log.Printf("listening on %v", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getUser(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, baseURL)
	if name == "" {
		users := getSystemUsers()
		b, err := json.Marshal(users)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v", err)
			return
		}
		fmt.Fprintf(w, "%s", b)

	} else {
		u, err := user.Lookup(name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "user not found: %s", name)
			return
		}
		b, err := json.Marshal(u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v", err)
			return
		}
		fmt.Fprintf(w, "%s", b)
	}
}

func getSystemUsers() []*user.User {
	var users []*user.User
	for i := 0; i < maxUid; i++ {
		u, err := user.LookupId(strconv.Itoa(i))
		if err == nil {
			users = append(users, u)
		}
	}
	return users
}
