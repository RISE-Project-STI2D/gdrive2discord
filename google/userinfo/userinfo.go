package userinfo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"../../google"
)

type response struct {
	Error       *google.ErrorResponse `json:"error"`
	DisplayName string                `json:"displayName"`
	Name        name                  `json:"name"`
	Emails      []email               `json:"emailAddresses"`
}

type name struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type email struct {
	Value string `json:"value"`
}

type UserInfo struct {
	DisplayName string `json:"displayName"`
	GivenName   string `json:"givenName"`
	FamilyName  string `json:"familyName"`
	Email       string `json:"emailAddresses"`
}

func GetUserInfo(client *http.Client, accessToken string) (*UserInfo, google.StatusCode, error) {
	u, _ := url.Parse("https://people.googleapis.com/v1/people/me")
	q := u.Query()
	q.Set("personFields", "names,emailAddresses")
	u.RawQuery = q.Encode()
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, google.CannotConnect, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	var deser = new(response)
	err = json.Unmarshal(body, &deser)
	if err != nil {
		return nil, google.CannotDeserialize, err
	}
	if deser.Error != nil {
		if deser.Error.Code == 401 {
			return nil, google.Unauthorized, errors.New(deser.Error.Message)
		}
		return nil, google.ApiError, errors.New(deser.Error.Message)
	}
	userInfo := &UserInfo{
		DisplayName: deser.DisplayName,
		GivenName:   deser.Name.GivenName,
		FamilyName:  deser.Name.FamilyName,
		Email:       deser.Emails[0].Value,
	}
	return userInfo, google.Ok, nil
}
