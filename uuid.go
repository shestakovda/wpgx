package wpgx

import (
	"encoding/hex"

	"github.com/satori/go.uuid"
)

func GUID() string {
	v4, _ := uuid.NewV4()
	return v4.String()
}

func UUID() string {
	v4, _ := uuid.NewV4()
	return hex.EncodeToString(v4.Bytes())
}
