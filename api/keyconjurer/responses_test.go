package keyconjurer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseMarshalJSON(t *testing.T) {
	type T struct {
		Foo, Bar string
	}

	data, _ := DataResponse(T{Foo: "Foo", Bar: "Qux"})
	b, _ := json.Marshal(data)

	require.Equal(t, `{"Success":true,"Message":"success","Data":{"Foo":"Foo","Bar":"Qux"}}`, string(b))
}

func TestResponseGetPayload(t *testing.T) {
	payload := `{"Success":true,"Message":"","Data":{"foo": "bar", "qux": "baz"}}`
	var response Response
	var data map[string]string
	var err error
	require.Error(t, response.GetPayload(&data))
	require.Error(t, response.GetError(&err))
	require.Nil(t, err)
	require.NoError(t, json.Unmarshal([]byte(payload), &response))
	require.NoError(t, response.GetPayload(&data))
	require.Error(t, response.GetError(&err))
	require.Nil(t, err)
	require.Equal(t, "bar", data["foo"])
	require.Equal(t, "baz", data["qux"])
}

func TestResponseGetError(t *testing.T) {
	payload := `{"Success":false,"Data":{"Code": "unspecified", "Message": "Something broke"}}`
	var response Response
	var data map[string]string
	var err error
	require.Error(t, response.GetPayload(&data))
	require.NoError(t, json.Unmarshal([]byte(payload), &response))
	require.Error(t, response.GetPayload(&data))
	require.NoError(t, response.GetError(&err))
	require.Equal(t, "Something broke (code: unspecified)", err.Error())
}
