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
	"encoding/json"
	"sync"
)

type Score struct {
	mutex		sync.Mutex
	results		map[string]map[string]int
}

func NewScoreFromPolls(polls map[string]Poll) (*Score) {

	s := &Score{
		mutex:   sync.Mutex{},
		results: map[string]map[string]int{},
	}

	for name,poll := range polls {
		s.results[name] = map[string]int{}
		for option := range poll.Options {
			s.results[name][option] = 0
		}
	}
	return s
}

func NewScoreFromJson(r []byte) (*Score) {

	s := &Score{
		mutex:   sync.Mutex{},
		results: map[string]map[string]int{},
	}

	json.Unmarshal(r, s.results)
	return s
}

func (s *Score) GetResultsInJson() []byte {

	r := s.GetResults()
	j, _ := json.Marshal(r)
	return j
}


func (s *Score) GetCopyResults() map[string]map[string]int {

		var copyResults = map[string]map[string]int{}

		s.mutex.Lock()
		defer s.mutex.Unlock()

		for key, mapV := range s.results {
			copyResults[key] = map[string]int{}
			for keyM, v := range mapV {
				copyResults[key][keyM] = v
			}
		}
		return copyResults
}

func (s *Score) GetResults() map[string]map[string]int {
	return s.results
}


func (s *Score) VoteFor(poll string, vote string) bool {

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, found := s.results[poll][vote]; found {
		s.results[poll][vote]++
		return true
	} else {
		return false
	}

}