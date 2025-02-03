package identity

import "fmt"

var (
	ErrRingCannotTrustSecretKey = fmt.Errorf("faws/identity: you cannot trust your own secret key; it is already trusted")
	ErrRingTooManyKeys          = fmt.Errorf("faws/identity: too many keys in keyring")
	ErrRingKeyNotFound          = fmt.Errorf("faws/identity: no ID found")
	ErrRingNoNametag            = fmt.Errorf("faws/identity: no nametag")
	ErrRingNametagInUse         = fmt.Errorf("faws/identity: nametag in use")
	ErrAbbreviationAmbiguous    = fmt.Errorf("faws/identity: ID abbreviation is ambiguous")
)
