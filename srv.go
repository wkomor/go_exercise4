package main

import (
	"io"
	"strconv"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"net/http"
)

func checkAccessToken(w http.ResponseWriter, r *http.Request) error {
	accessToken := r.Header.Get("AccessToken")
	if accessToken != "AAAbbb" {
		http.Error(w, "Invaid token", http.StatusUnauthorized)
		return errors.New("Invaid token")
	}
	return nil
}

func decodeRow(token xml.CharData) {
	var row Row
	xml.Unmarshal([]byte(token), &row)
	fmt.Println(row.Row.FirstName)
}

// Row struct
type Row struct {
	Row		RowElement `xml:"row"`
}


// RowElement struct
type RowElement struct {
	ID			int `xml:"id"`
    GUID		string `xml:"guid"`
    IsActive	bool `xml:"isActive"`
    Balance		string `xml:"balance"`
    Picture		string `xml:"picture"`
    Age			int `xml:"age"`
    EyeColor	string `xml:"eyeColor"`
    FirstName	string `xml:"first_name"`
    LastName	string `xml:"last_name"`
    Gender		string `xml:"gender"`
    Company		string `xml:"company"`
    Email		string `xml:"email"`
    Phone		string `xml:"phone"`
    Address		string `xml:"address"`
    About		string `xml:"about"`
    Registered	string `xml:"registered"`
    FavoriteFruit	string `xml:"favoriteFruit"`
}

	

// SearchServer dgfg
func SearchServer(w http.ResponseWriter, r *http.Request) {
	err := checkAccessToken(w, r)
	if err != nil {
		return
	}

	// limit := r.URL.Query().Get("limit")
	// offset := r.URL.Query().Get("offset")
	// query := r.URL.Query().Get("query")
	// orderField := r.URL.Query().Get("order_field")
	// orderBy := r.URL.Query().Get("order_by")
	
	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()

	// intLimit, err := strconv.Atoi(limit)
	if err != nil {
		panic(err)
	}
	// result := make([]string, intLimit)

	decoder := xml.NewDecoder(xmlFile)
	var ID string
	var firstName string
	var lastName string
	var user User
	// var users []User
	for {
		token, tokenErr := decoder.Token()
		if tokenErr != nil && tokenErr != io.EOF {
			fmt.Println("error happend", tokenErr)
			break
		} else if tokenErr == io.EOF {
			break
		}
		if token == nil {
			fmt.Println("t is nil break")
		}
		switch startElement := token.(type) {
			case xml.StartElement:
				if startElement.Name.Local == "row" {
					innerText, ok := token.(xml.CharData)
					if !ok {
						continue
					}
					decodeRow(innerText)
					
				}

				if startElement.Name.Local == "first_name" {
					if err := decoder.DecodeElement(&firstName, &startElement); err != nil {
						fmt.Println("error happend", err)
					}
				}
				if startElement.Name.Local == "last_name" {
					if err := decoder.DecodeElement(&lastName, &startElement); err != nil {
						fmt.Println("error happend", err)
					}
				}
				user.Id, _ = strconv.Atoi(ID)
				user.Name = firstName + " " + lastName
			}
		// fmt.Println(user)
	}
}

