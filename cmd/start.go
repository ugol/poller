// Copyright © 2018
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)


type Poll struct {
	Name            string
	Options         []string `json:"options"`
}

var (
	pollJson		string
	logDir     		string
	address			string

	//APP_ID			string
	//APP_KEY			string
	//APP_SECRET		string
	//APP_CLUSTER		string

	writeTimeout 	time.Duration
	readTimeout 	time.Duration
	idleTimeout 	time.Duration
	gracefulTimeout time.Duration

	poll *Poll
	results 		map[string]int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the poller server",
	Long: `Start the poller server:

poller start --address localhost:8080
poller start --address localhost:8080 --gracefulTimeout 1m
poller start --address localhost:8080 --gracefulTimeout 1m --readTimeout 30s`,
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func ResultsHandler(w http.ResponseWriter, r *http.Request) {


	result, err := json.Marshal(results)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
		log.Printf("Results: %v\n", results)
	}

}

func PollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	//client := pusher.Client{
	//	AppId:   APP_ID,
	//	Key:     APP_KEY,
	//	Secret:  APP_SECRET,
	//	Cluster: APP_CLUSTER,
	//	Secure:  true,
	//}
	//
	//data := map[string]string{"vote": vars["vote"]}
	//client.Trigger(poll.Name, "vote-event", data)

	results[vars["vote"]]++
	fmt.Fprintf(w, "You voted: %v\n", vars["vote"])
	log.Printf("Vote received: %v\n", vars["vote"])
}

func init() {

	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&logDir, "logdir", "tmp/poller", "")
	startCmd.Flags().StringVar(&address, "address", "localhost:8080", "Address to bind on")
	startCmd.Flags().StringVar(&pollJson, "pollJson", "polls/default.json", "Name of the JSON file describing the poll")

	//startCmd.Flags().StringVar(&APP_ID, "APP_ID", os.Getenv("APP_ID"), "Pusher APP_ID")
	//startCmd.Flags().StringVar(&APP_ID, "APP_KEY", os.Getenv("APP_KEY"), "Pusher APP_KEY")
	//startCmd.Flags().StringVar(&APP_ID, "APP_SECRET", os.Getenv("APP_SECRET"), "Pusher APP_SECRET")
	//startCmd.Flags().StringVar(&APP_ID, "APP_CLUSTER", os.Getenv("APP_CLUSTER"), "Pusher APP_CLUSTER")

	startCmd.Flags().DurationVar(&writeTimeout, "writeTimeout", time.Second * 15, "Write Timeout")
	startCmd.Flags().DurationVar(&readTimeout, "readTimeout", time.Second * 15, "Read Timeout")
	startCmd.Flags().DurationVar(&idleTimeout, "idleTimeout", time.Second * 60, "Idle Timeout")
	startCmd.Flags().DurationVar(&gracefulTimeout, "gracefulTimeout", time.Second * 15, "Graceful Timeout is the duration for which the server gracefully wait for existing connections to finish")

	results = make(map[string]int)
}

func start() {

	log.Println("Starting Poller server")


	file, err := os.Open(pollJson)
	if err != nil {
		log.Println(err)
		os.Exit(1)

	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&poll)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Printf("Using JSON \"%s\" poll description:\n %s \n", pollJson, poll)

	log.Printf("\n GracefulTimeout %s\n WriteTimeout %s\n ReadTimeout %s\n IdleTimeout %s\n", gracefulTimeout, writeTimeout, readTimeout, idleTimeout )

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("html"))))

	var baseUrl = fmt.Sprintf("/polls/%s", poll.Name)
	var voteUrl = fmt.Sprintf("/polls/%s/{vote}", poll.Name)

	r.HandleFunc(baseUrl, ResultsHandler).Methods("GET")
	r.HandleFunc(voteUrl, PollHandler).Methods("POST")

	http.Handle("/", r)


	srv := &http.Server{

		Addr:         address,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Printf("Waiting for connections at %s\n", address)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}