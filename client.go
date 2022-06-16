package ablearcher

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

func setupPutRequest(req *http.Request, datum *storageRecordWithKey) {
	req.Header.Set(headerVersion, "83")
	req.Header.Set("Content-Type", "text/html;charset=utf-8")
	zero := time.Time{}
	req.Header.Set(headerIfUnmodifiedSince, zero.Format(http.TimeFormat))
	req.Header.Set(headerAuthorization, fmt.Sprintf("Spring-83 Signature=%s", hex.EncodeToString(datum.auth)))
}
