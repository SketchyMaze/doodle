package license

import (
	"crypto/ecdsa"
	"fmt"
)

// Run-time configuration variables provided by the application.
const pemSigningKey = `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEMrqAMHjZ1dPlKwDOsiCSr5N3OSvnYKLM
efe2xD+5hJYrpvparRFnaMbMuqde4M6d6sCCKO8BHtfAzmyiQ/CD38zs9MiDsamy
FDYEEJu+Fqx482I7fIa5ZEE770+wWJ3k
-----END PUBLIC KEY-----`

var Signer *ecdsa.PublicKey

func init() {
	key, err := ParsePublicKeyPEM(pemSigningKey)
	if err != nil {
		fmt.Printf("license: failed to parse app keys: %s\n", err)
		return
	}

	Signer = key
}
