package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"
)

type user struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type problem struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
}

type snapshot struct {
	UserID       string `json:"user_id"`
	Revision     int64  `json:"revision"`
	CursorLine   int    `json:"cursor_line"`
	CursorColumn int    `json:"cursor_column"`
}

type match struct {
	ID             string     `json:"id"`
	State          string     `json:"state"`
	Problem        problem    `json:"problem"`
	RoundNumber    int        `json:"round_number"`
	RoundWinnerID  string     `json:"round_winner_id"`
	PlayerOneScore int        `json:"player_one_score"`
	PlayerTwoScore int        `json:"player_two_score"`
	CodeSnapshots  []snapshot `json:"code_snapshots"`
}

type submission struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Result    struct {
		Message     string `json:"message"`
		PassedTests int    `json:"passed_tests"`
		TotalTests  int    `json:"total_tests"`
		TestCases   []struct {
			Kind     string `json:"kind"`
			Status   string `json:"status"`
			Input    string `json:"input"`
			Expected string `json:"expected"`
			Actual   string `json:"actual"`
		} `json:"test_cases"`
	} `json:"result"`
}

type apiClient struct {
	base string
	http *http.Client
}

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:8088/api/v1", "API base URL")
	problemsDirectory := flag.String("problems", "problems", "problem catalog path")
	flag.Parse()

	aliceClient := newAPIClient(*baseURL)
	bobClient := newAPIClient(*baseURL)
	suffix := time.Now().UnixMilli()
	alice := register(aliceClient, fmt.Sprintf("alice_%d", suffix))
	bob := register(bobClient, fmt.Sprintf("bob_%d", suffix))

	var inviteResponse struct {
		Invitation struct {
			ID string `json:"id"`
		} `json:"invitation"`
	}
	aliceClient.do(http.MethodPost, "/invitations", map[string]any{"receiver_id": bob.ID}, &inviteResponse)
	var invitationState struct {
		Incoming *struct {
			ID string `json:"id"`
		} `json:"incoming"`
	}
	bobClient.do(http.MethodGet, "/invitations", nil, &invitationState)
	check(invitationState.Incoming != nil && invitationState.Incoming.ID == inviteResponse.Invitation.ID, "incoming invitation mismatch")

	var acceptResponse struct {
		Match match `json:"match"`
	}
	bobClient.do(http.MethodPost, "/invitations/"+inviteResponse.Invitation.ID+"/accept", nil, &acceptResponse)
	game := acceptResponse.Match
	check(game.Problem.Slug != "", "problem was not assigned")
	firstProblemID := game.Problem.ID
	source, err := os.ReadFile(filepath.Join(*problemsDirectory, game.Problem.Slug, "reference.go"))
	must(err)
	wrongSource, err := os.ReadFile(filepath.Join(*problemsDirectory, game.Problem.Slug, "starter.go"))
	must(err)

	aliceClient.do(http.MethodPut, "/matches/"+game.ID+"/code", map[string]any{
		"source_code":   string(source),
		"revision":      1,
		"cursor_line":   4,
		"cursor_column": 7,
	}, nil)
	var matchResponse struct {
		Match match `json:"match"`
	}
	bobClient.do(http.MethodGet, "/matches/"+game.ID, nil, &matchResponse)
	visible := false
	for _, code := range matchResponse.Match.CodeSnapshots {
		visible = visible || (code.UserID == alice.ID && code.Revision == 1 && code.CursorLine == 4 && code.CursorColumn == 7)
	}
	check(visible, "opponent did not receive editor snapshot and cursor")

	wrong := submitAndWait(aliceClient, game.ID, string(wrongSource))
	check(wrong.Status == "wrong_answer", fmt.Sprintf("wrong solution status %s: %s", wrong.Status, wrong.Result.Message))
	check(wrong.Result.Message == "Публичный пример 1 не пройден", "public-test feedback is not specific")
	check(len(wrong.Result.TestCases) >= 3, "public and hidden test results are missing")
	check(wrong.Result.TestCases[0].Kind == "public" && wrong.Result.TestCases[0].Status == "failed", "public test result mismatch")
	lastWrongTest := wrong.Result.TestCases[len(wrong.Result.TestCases)-1]
	check(lastWrongTest.Kind == "hidden" && lastWrongTest.Input == "" && lastWrongTest.Expected == "" && lastWrongTest.Actual == "", "hidden-test data leaked")

	// A long cold compilation can pause the room because the second client is idle.
	// Refresh both sessions before the accepted submission to exercise reconnect as well.
	aliceClient.do(http.MethodGet, "/matches/"+game.ID, nil, &matchResponse)
	bobClient.do(http.MethodGet, "/matches/"+game.ID, nil, &matchResponse)

	// The API enforces the interval using its own database clock. Waiting a
	// little over two seconds avoids boundary races with the client clock.
	time.Sleep(2100 * time.Millisecond)

	judged := submitAndWait(aliceClient, game.ID, string(source))
	check(judged.Status == "accepted", fmt.Sprintf("judge status %s: %s", judged.Status, judged.Result.Message))
	check(judged.Result.TotalTests > 0 && judged.Result.PassedTests == judged.Result.TotalTests, "accepted solution does not report all tests as passed")

	bobClient.do(http.MethodGet, "/matches/"+game.ID, nil, &matchResponse)
	finished := matchResponse.Match
	check(finished.State == "waiting_ready" && finished.RoundWinnerID == alice.ID, "round winner mismatch")
	check(finished.PlayerOneScore == 1 && finished.PlayerTwoScore == 0, "score was not incremented exactly once")

	aliceClient.do(http.MethodPost, "/matches/"+game.ID+"/ready", nil, nil)
	bobClient.do(http.MethodPost, "/matches/"+game.ID+"/ready", nil, &matchResponse)
	next := matchResponse.Match
	check(next.State == "active" && next.RoundNumber == 2, "next round did not start")
	check(next.Problem.ID != firstProblemID, "problem repeated in the next round")
	aliceClient.do(http.MethodPost, "/matches/"+game.ID+"/leave", nil, nil)

	output, _ := json.MarshalIndent(map[string]any{
		"alice":         alice.Username,
		"bob":           bob.Username,
		"match_id":      game.ID,
		"first_problem": game.Problem.Slug,
		"wrong_status":  wrong.Status,
		"judge_status":  judged.Status,
		"score":         fmt.Sprintf("%d:%d", finished.PlayerOneScore, finished.PlayerTwoScore),
		"next_problem":  next.Problem.Slug,
		"round":         next.RoundNumber,
	}, "", "  ")
	fmt.Println(string(output))
}

func submitAndWait(client *apiClient, matchID, source string) submission {
	var submitResponse struct {
		Submission submission `json:"submission"`
	}
	client.do(http.MethodPost, "/matches/"+matchID+"/submissions", map[string]any{
		"source_code": source,
	}, &submitResponse)
	judged := submitResponse.Submission
	deadline := time.Now().Add(60 * time.Second)
	for isPending(judged.Status) && time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		var response struct {
			Submission submission `json:"submission"`
		}
		client.do(http.MethodGet, "/submissions/"+judged.ID, nil, &response)
		judged = response.Submission
	}
	check(!isPending(judged.Status), "judge did not finish before timeout")
	return judged
}

func newAPIClient(base string) *apiClient {
	jar, err := cookiejar.New(nil)
	must(err)
	return &apiClient{base: base, http: &http.Client{Jar: jar, Timeout: 15 * time.Second}}
}

func register(client *apiClient, username string) user {
	var response struct {
		User user `json:"user"`
	}
	client.do(http.MethodPost, "/auth/register", map[string]any{
		"username": username,
		"password": "secret123",
	}, &response)
	return response.User
}

func (c *apiClient) do(method, path string, body any, target any) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		must(err)
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, c.base+path, reader)
	must(err)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(request)
	must(err)
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	must(err)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		panic(fmt.Sprintf("%s %s returned %d: %s", method, path, response.StatusCode, responseBody))
	}
	if target != nil && len(responseBody) > 0 {
		must(json.Unmarshal(responseBody, target))
	}
}

func isPending(status string) bool {
	return status == "queued" || status == "compiling" || status == "running"
}

func check(ok bool, message string) {
	if !ok {
		panic(errors.New(message))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
