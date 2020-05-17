package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const invalidToken = "invalid_token"
const timeoutError = "timeout_query"
const internalError = "fatal_query"
const badRequestError = "bad_request_query"
const badRequestUnknownError = "bad_request_unknown_query"
const invalidJSONError = "invalid_json_query"
const pageSize = 25

// TestCaseWithError struct
type TestCaseWithError struct {
	Request       SearchRequest
	URL           string
	AccessToken   string
	ErrorExact    string
	ErrorContains string
}

// TestCase struct
type TestCase struct {
	Request SearchRequest
}

func checkWrongQuery(w http.ResponseWriter, query string) error {
	switch query {
	case invalidToken:
		http.Error(w, "Invaid token", http.StatusUnauthorized)
		return errors.New("Invaid token")
	case timeoutError:
		time.Sleep(time.Second * 2)
		return nil
	case internalError:
		http.Error(w, "internal error", http.StatusInternalServerError)
		return errors.New("internal error")
	case badRequestError:
		http.Error(w, "bad request", http.StatusBadRequest)
		return errors.New("bad request")
	case badRequestUnknownError:
		resp, _ := json.Marshal(SearchErrorResponse{"UnknownError"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return errors.New("bad request unknown")
	case invalidJSONError:
		w.Write([]byte("invalid json"))
		return nil
	default:
		return nil
	}
}

// Document struct
type Document struct {
	Version string       `xml:"version"`
	Row     []RowElement `xml:"row"`
}

// RowElement struct
type RowElement struct {
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

// SearchServer dgfg
func SearchServer(w http.ResponseWriter, r *http.Request) {

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		fmt.Println("cant convert limit to int: ", err)
		return
	}
	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		fmt.Println("cant convert offset to int: ", err)
		return
	}
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")

	err = checkWrongQuery(w, query)
	if err != nil {
		return
	}

	if orderField == "" {
		orderField = "Name"
	}

	if orderField != "Id" && orderField != "Age" && orderField != "Name" {
		resp, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	xmlFile, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}

	document := &Document{}
	xml.Unmarshal(xmlFile, document)
	var users []User
	for _, user := range document.Row {
		users = append(users, User{
			Id:     user.ID,
			Name:   user.FirstName,
			Age:    user.Age,
			About:  user.About,
			Gender: user.Gender,
		})
	}


	users = users[offset:limit]

	jsonResponse, err := json.Marshal(users)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func TestStatusUnauthorized(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{Query: invalidToken})

	if err.Error() != "Bad AccessToken" {
		t.Error("Test AccessToken failed")
	}
}

func TestErrorResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}

	searchRequest := SearchRequest{
		Limit:  5,
		Offset: 0,
	}

	_, err := searchClient.FindUsers(searchRequest)
	if err != nil {
		t.Error("Dosn't work success request")
	}
	searchRequest.Limit = -1

	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "limit must be > 0" {
		t.Error("limit must be > 0")
	}

	searchRequest.Limit = 1
	searchRequest.Offset = -1
	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "offset must be > 0" {
		t.Error("offset must be > 0")
	}
}

func TestResultLimit(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}
	searchRequest := SearchRequest{
		Limit:  25,
		Offset: 1,
	}

	resp, _ := searchClient.FindUsers(searchRequest)
	if len(resp.Users) != pageSize {
		t.Error("Limit exceeded")
	}
}

func TestLimitFailed(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}

	limit := 7
	response, _ := searchClient.FindUsers(SearchRequest{Limit: limit})

	if limit != len(response.Users) {
		t.Error("Limit not true")
	}
}

func TestBadJson(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{Query: badRequestError})
	fmt.Println(err)
	if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Error("json test failed")
	}
}

func TestExceedLimit(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}

	response, _ := searchClient.FindUsers(SearchRequest{Limit: 26})

	if 25 != len(response.Users) {
		t.Error("Perelimit :(")
	}
}

func TestTimeoutError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}

	_, err := searchClient.FindUsers(SearchRequest{Query: timeoutError})

	if err == nil {
		t.Error("Timeout check failed")
	}
}

func TestUnknownError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}

	_, err := searchClient.FindUsers(SearchRequest{Query: badRequestUnknownError})

	if err == nil {
		t.Error("TestUnknownError failed")
	}
}

func TestInternalServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{Query: internalError})

	if err.Error() != "SearchServer fatal error" {
		t.Error("Internal error test failed")
	}
}

func TestInvalidJSONError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{Query: invalidJSONError})
	fmt.Println(err)
	if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Error("Internal error test failed")
	}
}

func TestUnknownAdressError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: "http://",
	}
	_, err := searchClient.FindUsers(SearchRequest{Query: invalidJSONError})
	if !strings.Contains(err.Error(), "unknown error") {
		t.Error("Unknown address test failed")
	}
}

func TestWrongOrderFieldError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()
	searchClient := &SearchClient{
		URL: testServer.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{OrderField: "order"})
	if err.Error() != "OrderFeld order invalid" {
		t.Error("Wrong order field test failed")
	}
}