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
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	resultAddress	string
	partialResults	map[string]map[string]int
)

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Start the results server",
	Long: `Start the results server:

poller results --address localhost:8082
poller results --address localhost:8082 --gracefulTimeout 1m
poller results --address localhost:8082 --gracefulTimeout 1m --readTimeout 30s`,
	Run: func(cmd *cobra.Command, args []string) {
		startResultServer()
	},
}

func ResultsHandler(w http.ResponseWriter, r *http.Request) {

	res := make(map[string]int)
	for _,option := range ThePoll.Options {
		res[option] = 0
	}

	for _, m := range partialResults {
		for k, v := range m {
			res[k] = res[k] + v
		}
	}

	result, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
		log.Printf("Results: %v\n", res)
	}

}

func init() {

	rootCmd.AddCommand(resultsCmd)

	resultsCmd.Flags().StringVar(&logDir, "logdir", "tmp/poller", "")
	resultsCmd.Flags().StringVar(&resultAddress, "resultAddress", "localhost:9090", "Address to bind on")
	resultsCmd.Flags().StringVar(&mcastAddress, "mcastAddress", "224.0.0.1:9999", "Multicast address used to broadcast the results")
	resultsCmd.Flags().StringVar(&pollJson, "pollJson", "polls/default.json", "Name of the JSON file describing the poll")

	resultsCmd.Flags().DurationVar(&writeTimeout, "writeTimeout", time.Second * 15, "Write Timeout")
	resultsCmd.Flags().DurationVar(&readTimeout, "readTimeout", time.Second * 15, "Read Timeout")
	resultsCmd.Flags().DurationVar(&idleTimeout, "idleTimeout", time.Second * 60, "Idle Timeout")
	resultsCmd.Flags().DurationVar(&gracefulTimeout, "gracefulTimeout", time.Second * 15, "Graceful Timeout is the duration for which the server gracefully wait for existing connections to finish")

}

func msgHandler(_ *net.UDPAddr, n int, b []byte) {

	appId := string(b[0:5])
	m := make(map[string]int)
	json.Unmarshal(b[5:n], &m)

	partialResults[appId] = m
	//log.Println(partialResults)

}

func getResultsFromMulticastUDP(a string, h func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	l.SetReadBuffer(8192)
	for {
		b := make([]byte, 8192)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}
}


func startResultServer() {

	log.Println("Starting Result server")


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

	log.Printf("Using JSON \"%s\" ThePoll description:\n %s \n", pollJson, ThePoll)

	partialResults = make(map[string]map[string]int)

	log.Printf("\n GracefulTimeout %s\n WriteTimeout %s\n ReadTimeout %s\n IdleTimeout %s\n", gracefulTimeout, writeTimeout, readTimeout, idleTimeout )

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("html"))))

	var baseUrl = fmt.Sprintf("/polls/%s", ThePoll.Name)

	r.HandleFunc(baseUrl, ResultsHandler).Methods("GET")

	http.Handle("/", r)

	srv := &http.Server{

		Addr:         resultAddress,
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

	log.Printf("Waiting for connections at %s\n", resultAddress)

	go func() {
		getResultsFromMulticastUDP(mcastAddress, msgHandler)
	}()

	log.Printf("Started receiving poll results on %s\n", mcastAddress)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}