// @APIVersion 1.0.0
// @APITitle Swagger IBM Cloud Provider API
// @APIDescription Swagger IBM Cloud Provider API
// @BasePath http://localhost:9080
// @Contact sakshiag@in.ibm.com

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fvbock/endless"
	"github.com/gorilla/mux"
	//	"github.com/terrform-schematics-demo/terraform-provider-ibm-api/utils"
	"github.com/terraform-provider-ibm-api/utils"
	mgo "gopkg.in/mgo.v2"
)

var staticContent = flag.String("staticPath", "./swagger/swagger-ui", "Path to folder with Swagger UI")
var apiurl = flag.String("api", "http://localhost", "The base path URI of the API service")

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	isJsonRequest := false

	if acceptHeaders, ok := r.Header["Accept"]; ok {
		for _, acceptHeader := range acceptHeaders {
			if strings.Contains(acceptHeader, "json") {
				isJsonRequest = true
				break
			}
		}
	}

	if isJsonRequest {
		w.Write([]byte(resourceListingJson))
	} else {
		http.Redirect(w, r, "/swagger-ui/", http.StatusFound)
	}
}

func ApiDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := strings.Trim(r.RequestURI, "/")
	fmt.Println("sakshiiii")
	if json, ok := apiDescriptionsJson[apiKey]; ok {
		t, e := template.New("desc").Parse(json)
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t.Execute(w, *apiurl)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	var port int
	flag.IntVar(&port, "p", 9080, "Port on which this server listens")
	flag.Parse()
	r := mux.NewRouter()

	r.HandleFunc("/", IndexHandler)

	//http.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir(*staticContent))))

	r.PathPrefix("/swagger-ui").Handler(http.StripPrefix("/swagger-ui", http.FileServer(http.Dir(*staticContent))))

	for apiKey := range apiDescriptionsJson {
		log.Println("sdsadsadsada", apiKey)
		r.HandleFunc("/"+apiKey, ApiDescriptionHandler)
	}

	r.HandleFunc("/v1/configuration", utils.ConfHandler(session)).Methods("POST")

	r.HandleFunc("/v1/configuration/{repo_name}", utils.ConfDeleteHandler).Methods("DELETE")

	r.HandleFunc("/v1/configuration/{repo_name}/plan", utils.PlanHandler(session)).Methods("POST")

	r.HandleFunc("/v1/configuration/{repo_name}/show", utils.ShowHandler(session)).Methods("POST")

	r.HandleFunc("/v1/configuration/{repo_name}/apply", utils.ApplyHandler(session)).Methods("POST")

	r.HandleFunc("/v1/configuration/{repo_name}/destroy", utils.DestroyHandler(session)).Methods("POST")

	r.HandleFunc("/v1/configuration/{repo_name}/{action}/{actionID}/log", utils.LogHandler).Methods("GET")

	r.HandleFunc("/v1/configuration/{repo_name}/{action}/{actionID}/status", utils.StatusHandler(session)).Methods("GET")

	r.HandleFunc("/v1/configuration/{repo_name}/{action}/{log_file}", utils.ViewLogHandler)

	r.HandleFunc("/v1/configuration/{repo_name}/{action}", utils.GetActionDetailsHandler(session)).Methods("GET")

	fmt.Println("Server will listen at port", port)
	muxWithMiddlewares := http.TimeoutHandler(r, time.Second*60, "Timeout!")
	err = endless.ListenAndServe(fmt.Sprintf(":%d", port), muxWithMiddlewares)
	if err != nil {
		fmt.Printf("Couldn't start the server %v", err)
	}
}

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()
	c := session.DB("action").C("actionDetails")

	index := mgo.Index{
		Key:        []string{"actionid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}
