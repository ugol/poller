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
	"testing"
)

func TestScoreMap(t *testing.T) {

	p1 := Poll{
		PollDescription:	"",
		Options: 			map[string]string{"one":"one", "two":"two", "three":"three"},
	}
	p2 := Poll{
		PollDescription:	"",
		Options: 			map[string]string{"four":"four", "five":"five", "six":"six"},
	}
	polls := map[string]Poll{"1":p1, "2":p2}

	score:= NewScoreFromPolls(polls)

	score.VoteFor("1", "three")
	score.VoteFor("1", "three")
	score.VoteFor("2", "four")
	score.VoteFor("2", "five")

	valid := score.VoteFor("2", "six")
	invalid := score.VoteFor("2", "seven")

	if !valid {
		t.Errorf("Six should be a valid vote, but the vote returned: %v", valid)
	}

	if invalid {
		t.Errorf("Seven shouldn't be a valid vote, but the vote returned: %v", invalid)
	}

	scoreResults := score.GetResults()
	scoreCopyResults := score.GetCopyResults()

	if scoreResults["1"]["one"] != scoreCopyResults["1"]["one"] {
		t.Errorf("Same values expected, but are different %v == %v", scoreResults["1"]["one"], scoreCopyResults["1"]["one"])
	}

	if scoreResults["1"]["two"] != scoreCopyResults["1"]["two"] {
		t.Errorf("Same values expected, but are different %v == %v", scoreResults["1"]["two"], scoreCopyResults["1"]["two"])
	}

	if &scoreResults == &scoreCopyResults {
		t.Errorf("Different address expected, but are the same %p == %p", scoreResults, scoreCopyResults)
	}

	subOriginal1 := scoreResults["1"]
	s1 := &subOriginal1

	subOriginal2 := scoreResults["2"]
	s2 := &subOriginal2

	subCopy1 := scoreCopyResults["1"]
	q1 := &subCopy1

	subCopy2 := scoreCopyResults["2"]
	q2 := &subCopy2

	if s1 == q1 {
		t.Errorf("Different address expected, but are the same %p == %p", s1, q1)
	}

	if s2 == q2 {
		t.Errorf("Different address expected, but are the same %p == %p", s2, q2)
	}

}

