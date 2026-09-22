package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const PhoneIndex = "phone_numbers"

var phoneIndexMapping = []byte(`{
  "mappings": {
    "dynamic": "strict",
    "properties": {
      "id": {"type":"keyword"},
      "e164": {"type":"keyword"},
      "country_code": {"type":"keyword"},
      "calling_code": {"type":"keyword"},
      "national_number": {"type":"keyword"},
      "number_type": {"type":"keyword"},
      "verification_status": {"type":"keyword"},
      "seo_status": {"type":"keyword"},
      "spam_score": {"type":"float"},
      "report_count": {"type":"long"},
      "search_count": {"type":"long"},
      "data_quality_score": {"type":"float"},
      "last_seen_at": {"type":"date"}
    }
  }
}`)

func (c *Client) EnsurePhoneIndex(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL+"/"+PhoneIndex, nil)
	if err != nil { return err }
	resp, err := c.http.Do(req)
	if err != nil { return err }
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK { return nil }
	if resp.StatusCode != http.StatusNotFound { return fmt.Errorf("opensearch index check returned %s", resp.Status) }

	req, err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/"+PhoneIndex, strings.NewReader(string(phoneIndexMapping)))
	if err != nil { return err }
	req.Header.Set("Content-Type","application/json")
	resp, err = c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("create phone index returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func MarshalPhoneDocument(v any) ([]byte,error) { return json.Marshal(v) }
