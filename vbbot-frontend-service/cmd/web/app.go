package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

const webPort = "3000"

type RequestPayload struct {
	Action   string          `json:"action"`
	Register RegisterPayload `json:"register,omitempty"`
}

type RegisterPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "index.page.gohtml")
	})
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		render(w, "auth.page.gohtml")
	})
	http.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {
		cook, err := r.Cookie("Auth")
		if err != nil {
			render(w, "register.failed.gohtml")
			return
		}
		if cook.Value == "true" {
			render(w, "account.page.gohtml")
		}
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			render(w, "register.page.gohtml")
		} else if r.Method == "POST" {
			r.ParseForm()
			requestPayload := RequestPayload{
				Action: r.Form["action"][0],
				Register: RegisterPayload{
					Username: r.Form["username"][0],
					Email:    r.Form["email"][0],
					Password: r.Form["password"][0],
				},
			}
			jsonData, err := json.MarshalIndent(requestPayload, "", "\t")
			if err != nil {
				fmt.Println(err)
				render(w, "register-failed.page.gohtml")
				return
			}
			request, err := http.NewRequest("POST", "http://localhost:8001/handle", bytes.NewBuffer(jsonData))
			client := http.Client{}
			response, err := client.Do(request)
			if err != nil {
				fmt.Println(err)
				authC := &http.Cookie{
					Name:   "Auth",
					Value:  "false",
					MaxAge: 300,
				}
				http.SetCookie(w, authC)
				render(w, "register-failed.page.gohtml")
				return
			}
			defer response.Body.Close()
			if response.StatusCode == http.StatusAccepted {
				fmt.Println("ok")
				authC := &http.Cookie{
					Name:   "Auth",
					Value:  "true",
					MaxAge: 300,
				}
				http.SetCookie(w, authC)
				http.Redirect(w, r, "/account", http.StatusSeeOther)
				//render(w, "account.page.gohtml")
			} else {
				authC := &http.Cookie{
					Name:   "Auth",
					Value:  "false",
					MaxAge: 300,
				}
				http.SetCookie(w, authC)
				render(w, "register-failed.page.gohtml")
			}
		}
	})
	fs := http.FileServer(http.Dir("./ui/static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	fmt.Printf("Start server on port: %s\n", webPort)
	err := http.ListenAndServe(fmt.Sprintf(":%s", webPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func render(w http.ResponseWriter, t string) {
	path, _ := os.Getwd()
	fmt.Println(path)

	partials := []string{
		"./ui/html/base.layout.gohtml",
		"./ui/html/head.partial.gohtml",
		"./ui/html/header.partial.gohtml",
		"./ui/html/footer.partial.gohtml",
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("./ui/html/%s", t))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
