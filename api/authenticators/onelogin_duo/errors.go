package oneloginduo

import "errors"

var ErrorAssumingRole = errors.New("Error assuming role")
var ErrorCannotFindCertsUrl = errors.New("Cannot find certs_url")
var ErrorCannotFindSid = errors.New("Cannot find sid")
var ErrorDuoCommunication = errors.New("Unable to communicate with Duo")
var ErrorDuoMfaNotAllow = errors.New("MFA was not allowed")
var ErrorDuoArgsError = errors.New("There was an error parsing arguments for Duo push request")
var ErrorDuoPushError = errors.New("There was an error sending a Duo push request")
var ErrorHttpBodyError = errors.New("Unable to read http response body")
var ErrorJsonMarshalError = errors.New("Unable to marshal json")
var ErrorJsonUnmarshalError = errors.New("Unable to unmarshal json")
var ErrorSamlAssertionHasTooManyRoles = errors.New("SAML assertion has too many roles")
var ErrorUnableToDecode = errors.New("Unable to base64 decode string")
var ErrorUnableToDecrypt = errors.New("Unable to decrypt data")
var ErrorUnableToEncrypt = errors.New("Unable to encrypt data")
var ErrorUnableToFindPreferredDevice = errors.New("Unable to find preferred device")
var ErrorUnableToGetSamlAssertion = errors.New("Unable to get SAML assertion")
var ErrorUnableToParseSamlAssertion = errors.New("Unable to parse SAML assertion")
