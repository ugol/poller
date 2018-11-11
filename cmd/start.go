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
	"html/template"
	"log"
	"net"
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
	pollerAddress	string
	mcastAddress	string

	APP_ID			string

	writeTimeout 	time.Duration
	readTimeout 	time.Duration
	idleTimeout 	time.Duration
	gracefulTimeout time.Duration

	mcastInterval	time.Duration

	ThePoll 		*Poll
	results 		map[string]int
	vote  =			template.Must(template.ParseFiles("html/vote.html"))

)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the poller server",
	Long: `Start the poller server:

poller start --address localhost:8080
poller start --address localhost:8080 --gracefulTimeout 1m
poller start --address localhost:8080 --gracefulTimeout 1m --readTimeout 30s`,
	Run: func(cmd *cobra.Command, args []string) {
		startPollServer()
	},
}

func PollHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		log.Printf("Serving POST %v to %s\n", r.RequestURI, r.RemoteAddr  )

		vars := mux.Vars(r)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		vote := vars["vote"]

		if _, found := results[vote]; found {
			results[vote]++
			fmt.Fprintf(w, "You voted: %v\n", vote)
			log.Printf("Vote received: %v\n", vote)

		} else {
			fmt.Fprintf(w, "Invalid voted received: %v\n", vote)
			log.Printf("Invalid vote: %v\n", vote)
		}
	} else if r.Method == http.MethodGet {
		log.Printf("Serving GET %v to %s\n", r.RequestURI, r.RemoteAddr  )
		err := vote.ExecuteTemplate(w,"vote.html", ThePoll)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}


}

func init() {

	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&logDir, "logdir", "tmp/poller", "")
	startCmd.Flags().StringVar(&pollerAddress, "pollerAddress", "localhost:8080", "Address to bind on")
	startCmd.Flags().StringVar(&mcastAddress, "mcastAddress", "224.0.0.1:9999", "Multicast address used to broadcast the results")
	startCmd.Flags().StringVar(&pollJson, "pollJson", "polls/default.json", "Name of the JSON file describing the poll")

	startCmd.Flags().StringVar(&APP_ID, "APP_ID", os.Getenv("APP_ID"), "APP_ID is an unique identifier that must be used when scaling poller horizontally. Only the last 5 bytes are significant")

	startCmd.Flags().DurationVar(&mcastInterval, "mcastInterval", time.Second * 1, "Interval to multicast the results")

	startCmd.Flags().DurationVar(&writeTimeout, "writeTimeout", time.Second * 15, "Write Timeout")
	startCmd.Flags().DurationVar(&readTimeout, "readTimeout", time.Second * 15, "Read Timeout")
	startCmd.Flags().DurationVar(&idleTimeout, "idleTimeout", time.Second * 60, "Idle Timeout")
	startCmd.Flags().DurationVar(&gracefulTimeout, "gracefulTimeout", time.Second * 15, "Graceful Timeout is the duration for which the server gracefully wait for existing connections to finish")

}

func broadcastResults() {
	addr, err := net.ResolveUDPAddr("udp", mcastAddress)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	for {

		r, err := json.Marshal(results)

		if err != nil {
			log.Println(err)
		} else {
			c.Write([]byte(fmt.Sprintf("%s%s", APP_ID[len(APP_ID)-5:], r)))
		}

		time.Sleep(mcastInterval)
	}
}

func startPollServer() {

	log.Println("Starting Poller server")

	if len(APP_ID)<5 {
		log.Printf("APP_ID must be 5 bytes or greater, but is %d bytes. ('%s')\n", len(APP_ID), APP_ID)
		os.Exit(1)
	}

	file, err := os.Open(pollJson)
	if err != nil {
		log.Println(err)
		os.Exit(1)

	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&ThePoll)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Printf("Using JSON \"%s\" poll description:\n %s \n", pollJson, ThePoll)

	results = make(map[string]int)

	for _,option := range ThePoll.Options {
		results[option] = 0
	}

	log.Printf("\n GracefulTimeout %s\n WriteTimeout %s\n ReadTimeout %s\n IdleTimeout %s\n", gracefulTimeout, writeTimeout, readTimeout, idleTimeout )

	r := mux.NewRouter()

	//r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("html"))))

	var voteUrl = fmt.Sprintf("/polls/%s/leaveyourvote", ThePoll.Name)
	var votesUrl = fmt.Sprintf("/polls/%s/{vote}", ThePoll.Name)

	r.HandleFunc(voteUrl, PollHandler).Methods("GET")
	r.HandleFunc(votesUrl, PollHandler).Methods("POST")

	http.Handle("/", r)


	srv := &http.Server{

		Addr:         pollerAddress,
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

	log.Printf("Waiting for connections at %s\n", pollerAddress)

	go func() {
		broadcastResults()
	}()

	log.Printf("Started broadcasting poll results on %s using APP_ID '%s' every %s\n", mcastAddress, APP_ID, mcastInterval)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}