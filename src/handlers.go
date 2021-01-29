package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (s *Server) verifycredentials() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle VerifyCredentials in IS with SCIM Has Been Called!")
		//get JSON payload
		user := User{}
		err := json.NewDecoder(r.Body).Decode(&user)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println(err.Error())
			return
		}

		if user.KeySecret == "" {
			invalidUser := UserResponse{}
			invalidUser.Message = "Bad Json provided! No application authorization token provided..."
			js, jserr := json.Marshal(invalidUser)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to validate user credentials...")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client := &http.Client{}
		data := url.Values{}
		data.Set("grant_type", "password")
		data.Add("username", user.Username)
		data.Add("password", user.Password)
		//data.Add("scope", user.Scopes[0].Scope)
		test := " "
		for _, element := range user.Scopes {
			// index is the index where we are
			// element is the element from someSlice for where we are
			test = test + element.Scope + " "
		}
		fmt.Println(test)
		data.Add("scope", test)

		fmt.Println(data)

		req, err := http.NewRequest("POST", "https://"+config.APIMHost+"/token", bytes.NewBufferString(data.Encode()))

		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", "Basic "+user.KeySecret)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == 400 {
			invalidUser := UserResponse{}
			invalidUser.Message = "Invalid user credentials for application / Invalid application authorization token."
			bodyText, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(bodyText)
			js, jserr := json.Marshal(invalidUser)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to verify user credentials when incorrect credential details for the application were recieved...")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s\n", bodyText)
		identityServerResponse := TokenResponse{}
		validUser := UserResponse{}
		err = json.Unmarshal(bodyText, &identityServerResponse)

		if identityServerResponse.Accesstoken == "" {
			validUser.Message = "User found but invalid application authorization token provided..."
		} else {
			validUser.Message = "User credentials successfully validated!"
		}

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding validate credentials response...")
			return
		}

		validUser.Accesstoken = identityServerResponse.Accesstoken
		validUser.Refreshtoken = identityServerResponse.Refreshtoken
		validUser.Scopes = identityServerResponse.Scopes
		js, jserr := json.Marshal(validUser)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to validate user credentials...")
			return
		}

		//return back to Front-End user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)

	}
}

func (s *Server) handleregisteruser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle Register User in IS with SCIM Has Been Called!")

		regUser := RegisterUser{}
		err := json.NewDecoder(r.Body).Decode(&regUser)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Improper registration details provided")
			return
		}
		//fmt.Println(regUser.Name + "\n" + regUser.Surname + "\n" + regUser.Password + "\n" + regUser.Email + "\n" + regUser.Username)
		/*if regUser.KeySecret != config.KeySecret {
			keyErrorByte, _ := json.Marshal("Resource accessed without the correct key and secret!")
			w.WriteHeader(500)
			w.Write(keyErrorByte)
			fmt.Println("Resource accessed without the correct key and secret!")
			return
		}*/

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client := &http.Client{}
		//var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + regUser.Surname + `" ,"givenName":"` + regUser.Name + `"},"userName":"` + regUser.Username + `","password":"` + regUser.Password + `","emails":[{"primary":true,"value":"` + regUser.Email + `","type":"home"},{"value":"` + regUser.Email + `","type":"work"}]}`)
		var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + regUser.Surname + `" ,"givenName":"` + regUser.Name + `"},"userName":"` + regUser.Username + `","password":"` + regUser.Password + `","emails":[{"primary":true,"value":"` + regUser.Email + `","type":"home"},{"value":"` + regUser.Email + `","type":"work"}]}`)

		req, err := http.NewRequest("POST", "https://"+config.ISHost+"/wso2/scim/Users", data)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "admin")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == 409 {
			userExists := RegisterUserResponse{}
			userExists.UserCreated = "false"
			userExists.Username = "n/a"
			userExists.UserID = "00000000-0000-0000-0000-000000000000"
			userExists.Message = "This Username Already Exists!"

			js, jserr := json.Marshal(userExists)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to register user when that user already exists.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s\n", bodyText)

		var identityServerResponse IdentityServerResponse

		err = json.Unmarshal(bodyText, &identityServerResponse)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}
		var regUserResponse RegisterUserResponse

		regUserResponse.Message = "User successfully Registered!"
		regUserResponse.UserCreated = "true"
		regUserResponse.UserID = identityServerResponse.ID
		regUserResponse.Username = identityServerResponse.Username

		js, jserr := json.Marshal(regUserResponse)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to register user")
			return
		}

		/*requestByte, _ := json.Marshal(regUser)
		reqToUM, respErr := http.Post("http://"+config.UMHost+":"+config.UMPort+"/user", "application/json", bytes.NewBuffer(requestByte))

		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with User Manager service endpoint for request to register")
			return
		}
		if reqToUM.StatusCode != 200 {
			fmt.Fprint(w, "Request to DB can't be completed...")
			fmt.Println("Unable to process registration")
		}
		if reqToUM.StatusCode == 500 {
			w.WriteHeader(500)

			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "Request to DB can't be completed..."+bodyString)
			fmt.Println("Request to DB can't be completed..." + bodyString)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Registration is not able to be completed by internal error")
			return
		}

		//close the request
		defer reqToUM.Body.Close()

		var registerResponse RegisterUserResponse

		//decode request into decoder which converts to the struct
		decoder := json.NewDecoder(reqToUM.Body)

		err = decoder.Decode(&registerResponse)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}
		js, jserr := json.Marshal(registerResponse)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to register user")
			return
		}

		if registerResponse.UserCreated == "false" {
			client := &http.Client{}
			req, err := http.NewRequest("DELETE", "https://auth.studymoney.co.za:9445/wso2/scim/Users/"+identityServerResponse.ID, nil)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.SetBasicAuth("admin", "admin")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			bodyText, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Printf("%s\n", bodyText)
		}*/

		//return back to Front-End user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)

	}
}

func (s *Server) handleassigngroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//regUser := RegisterUser{}
		//err := json.NewDecoder(r.Body).Decode(&regUser)
		fmt.Println("Handle assign user to group function has been called!")
		getUsername := r.URL.Query().Get("userName")
		getGroupName := r.URL.Query().Get("groupName")

		//Check if user and group names were provided in URL

		if getUsername == "" {
			w.WriteHeader(500)
			fmt.Fprint(w, "Username not properly provided in URL")
			fmt.Println("Username not properly provided in URL")
			return
		}

		if getGroupName == "" {
			w.WriteHeader(500)
			fmt.Fprint(w, "Groupname not properly provided in URL")
			fmt.Println("Groupname not properly provided in URL")
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		client := &http.Client{}
		//var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + regUser.Surname + `" ,"givenName":"` + regUser.Name + `"},"userName":"` + regUser.Username + `","password":"` + regUser.Password + `","emails":[{"primary":true,"value":"` + regUser.Email + `","type":"home"},{"value":"` + regUser.Email + `","type":"work"}]}`)
		req, err := http.NewRequest("GET", "https://"+config.ISHost+"/wso2/scim/Users?filter=userName+Eq+%22"+getUsername+"%22", nil)
		//req, respErr := http.Get("https://" + config.ISHost + ":9445/wso2/scim/Users?filter=userName+Eq+%22" + getUsername + "%22")

		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "admin")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == 409 {
		}

		var userResponse getUserIDResponse

		err = json.Unmarshal(bodyText, &userResponse)

		req, err = http.NewRequest("GET", "https://"+config.ISHost+"/wso2/scim/Groups?filter=displayName+Eq+%22"+getGroupName+"%22", nil)

		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "admin")
		resp, err = client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		bodyText, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var groupResponse getGroupIDResponse

		err = json.Unmarshal(bodyText, &groupResponse)

		fmt.Println("user ID: " + userResponse.Resources[0].ID)
		fmt.Println("Username: " + userResponse.Resources[0].Username)
		fmt.Println("group ID: " + groupResponse.Resources[0].ID)
		fmt.Println("Group Name: " + groupResponse.Resources[0].DisplayName)
		//var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + regUser.Surname + `" ,"givenName":"` + regUser.Name + `"},"userName":"` + regUser.Username + `","password":"` + regUser.Password + `","emails":[{"primary":true,"value":"` + regUser.Email + `","type":"home"},{"value":"` + regUser.Email + `","type":"work"}]}`)

		var data = strings.NewReader(`{"displayName": "` + groupResponse.Resources[0].DisplayName + `","members":[{"value":"` + userResponse.Resources[0].ID + `","display": "` + userResponse.Resources[0].Username + `"}]}`)
		fmt.Println(data)
		req, err = http.NewRequest("PATCH", "https://"+config.ISHost+"/wso2/scim/Groups/"+groupResponse.Resources[0].ID, data)

		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "admin")
		resp, err = client.Do(req)
		if err != nil {
			var addResponse addToGroupResponse

			addResponse.Message = err.Error()
			addResponse.Success = "False"
			js, jserr := json.Marshal(addResponse)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to register user")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		var addResponse addToGroupResponse

		addResponse.Message = "Username: " + userResponse.Resources[0].Username + " successfully added to Group/Role: " + groupResponse.Resources[0].DisplayName
		addResponse.Success = "True"
		js, jserr := json.Marshal(addResponse)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to register user")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)

	}
}
