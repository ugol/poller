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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Poll struct {
	PollDescription string
	Options         map[string]string
}

var (
	pollJson        string
	logDir          string
	pollerAddress   string
	mcastAddress    string
	APP_ID          string
	writeTimeout    time.Duration
	readTimeout     time.Duration
	idleTimeout     time.Duration
	gracefulTimeout time.Duration
	mcastInterval   time.Duration
	cookieDuration	time.Duration
	Polls           map[string]Poll
	score           *Score
)

var (

	totalVotes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poller_total_votes",
		Help: "The total number of processed votes",
	})

	validVotes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poller_valid_votes",
		Help: "The total number of valid votes",
	})

	invalidVotes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poller_invalid_votes",
		Help: "The total number of invalid votes",
	})

	votesRetried = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poller_retried_votes",
		Help: "The total number of votes receiving an 'already voted' answer",
	})

)


var (
	vote = template.Must(template.ParseFiles("templates/vote.template"))

	startCmd = &cobra.Command{
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
)

func PollHandler(w http.ResponseWriter, r *http.Request) {

	poll := strings.Split(r.RequestURI, "/")[2]

	if r.Method == http.MethodPost {
		hasVoted, _ := r.Cookie("poller")
		log.Printf("Serving POST %v to %s\n", r.RequestURI, r.RemoteAddr)
		log.Printf("Cookie: %v\n", hasVoted)

		vars := mux.Vars(r)
		vote := vars["vote"]
		totalVotes.Inc()

		if hasVoted != nil && hasVoted.Value == poll {
			fmt.Fprint(w, "You have already voted for this poll\n")
			fmt.Fprint(w, "<br><a href=\"../../../static/results.html?poll=%v\">Go to results</a>", poll)
			log.Print("You have already voted for this poll\n")
			votesRetried.Inc()
		} else {
			if score.VoteFor(poll, vote) {
				expiration := time.Now().Add(cookieDuration)
				voted := http.Cookie{Name: "poller", Value: poll, Expires: expiration}
				http.SetCookie(w, &voted)
				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "You voted: %v\n", vote)
				fmt.Fprintf(w, "<br><a href=\"../../../static/results.html?poll=%v\">Go to results</a>", poll)
				log.Printf("Vote received: %v\n", vote)
				validVotes.Inc()

			} else {
				fmt.Fprintf(w, "Invalid voted received: %v\n", vote)
				log.Printf("Invalid vote: %v\n", vote)
				invalidVotes.Inc()
			}
		}

	} else if r.Method == http.MethodGet {
		log.Printf("Serving GET %v to %s\n", r.RequestURI, r.RemoteAddr)
		err := vote.ExecuteTemplate(w, "vote.template", Polls[poll])
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
	startCmd.Flags().DurationVar(&mcastInterval, "mcastInterval", time.Second*1, "Interval to multicast the results")
	startCmd.Flags().DurationVar(&writeTimeout, "writeTimeout", time.Second*15, "Write Timeout")
	startCmd.Flags().DurationVar(&readTimeout, "readTimeout", time.Second*15, "Read Timeout")
	startCmd.Flags().DurationVar(&idleTimeout, "idleTimeout", time.Second*60, "Idle Timeout")
	startCmd.Flags().DurationVar(&cookieDuration, "cookieDuration", time.Second*60*2, "Cookie duration: can't vote again if you have this cookie set")
	startCmd.Flags().DurationVar(&gracefulTimeout, "gracefulTimeout", time.Second*15, "Graceful Timeout is the duration for which the server gracefully wait for existing connections to finish")

}

func broadcastResults() {
	addr, err := net.ResolveUDPAddr("udp", mcastAddress)

	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	for {
		r := score.GetResultsInJson()
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

	if len(APP_ID) < 5 {
		log.Printf("APP_ID must be 5 bytes or greater, but is %d bytes. ('%s')\n", len(APP_ID), APP_ID)
		os.Exit(1)
	}

	file, err := os.Open(pollJson)
	if err != nil {
		log.Println(err)
		os.Exit(1)

	}

	defer file.Close()
	err = json.NewDecoder(file).Decode(&Polls)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Printf("Using JSON \"%s\" poll description:\n %s \n", pollJson, Polls)
	score = NewScoreFromPolls(Polls)
	log.Printf("\n GracefulTimeout %s\n WriteTimeout %s\n ReadTimeout %s\n IdleTimeout %s\n", gracefulTimeout, writeTimeout, readTimeout, idleTimeout)
	r := mux.NewRouter()
	for name := range Polls {
		var voteUrl = fmt.Sprintf("/polls/%s/leaveyourvote", name)
		var votesUrl = fmt.Sprintf("/polls/%s/{vote}", name)
		r.HandleFunc(voteUrl, PollHandler).Methods("GET")
		r.HandleFunc(votesUrl, PollHandler).Methods("POST")
	}

	r.Path("/metrics").Handler(promhttp.Handler())


	http.Handle("/", r)
	srv := &http.Server{
		Addr:         pollerAddress,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      r,
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

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error %v during shutdown\n", err)
		os.Exit(1)
	} else {
		log.Println("Shutting down")
		os.Exit(0)
	}

}
