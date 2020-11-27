package keyconjurer

import (
	"encoding/json"
	"errors"
	"strconv"
)

type UserDataAccount struct {
	AccountID   string `json:"id"`
	AccountName string `json:"name"`
}

func (uda *UserDataAccount) UnmarshalJSON(b []byte) error {
	// Older versions of KeyConjurer would store IDs in int64 format; we use this method to support decoding that
	var old struct {
		AccountID   json.RawMessage `json:"id"`
		AccountName string          `json:"name"`
	}

	if err := json.Unmarshal(b, &old); err != nil {
		return err
	}

	uda.AccountName = old.AccountName

	var t1 int64
	// If we successfully parsed a string, great
	if err := json.Unmarshal(old.AccountID, &uda.AccountID); err == nil {
		return nil
	}

	if uda.AccountID != "" {
		return nil
	}

	if err := json.Unmarshal(old.AccountID, &t1); err == nil {
		uda.AccountID = strconv.FormatInt(t1, 10)
		return nil
	}

	return errors.New("unable to parse AccountID")
}

func (uda UserDataAccount) ID() string {
	return uda.AccountID
}

func (uda UserDataAccount) Name() string {
	return uda.AccountName
}

// Device is an artifact from older times we should be able to
//  remove this after April 1 , 2019
type Device struct {
}
