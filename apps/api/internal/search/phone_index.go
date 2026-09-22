package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	PhoneIndex = "phone_numbers"
	PhoneIndexV1 = "phone_numbers_v1"
)

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
      "last_seen_at": {"type":"date"},
      "display_name": {"type":"text","fields":{"keyword":{"type":"keyword"}}}
    }
  }
}`)

func (c *Client) EnsurePhoneIndex(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL+"/"+PhoneIndexV1, nil)
	if err != nil { return err }
	resp, err := c.http.Do(req)
	if err != nil { return err }
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK { return nil }
	if resp.StatusCode != http.StatusNotFound { return fmt.Errorf("opensearch index check returned %s", resp.Status) }

	req, err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/"+PhoneIndexV1, strings.NewReader(string(phoneIndexMapping)))
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

func (c *Client) SwapPhoneAlias(ctx context.Context) error {
	body := `{"actions":[{"remove":{"index":"phone_numbers_*","alias":"`+PhoneIndex+`","must_exist":false}},{"add":{"index":"`+PhoneIndexV1+`","alias":"`+PhoneIndex+`"}}]}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/_aliases", strings.NewReader(body))
	if err != nil { return err }
	req.Header.Set("Content-Type","application/json")
	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("swap phone alias returned %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}
	return nil
}

func MarshalPhoneDocument(v any) ([]byte,error) { return json.Marshal(v) }

type PhoneSearchHit struct {
 ID string `json:"id"`
 E164 string `json:"e164"`
 DisplayName string `json:"display_name,omitempty"`
 VerificationStatus string `json:"verification_status"`
 SpamScore float64 `json:"spam_score"`
 SearchCount int64 `json:"search_count"`
}
func (c *Client) SearchPhones(ctx context.Context,q,country string,limit int)([]PhoneSearchHit,error){
 if limit<1||limit>50{limit=20}
 query:=map[string]any{"size":limit,"query":map[string]any{"bool":map[string]any{"must":[]any{map[string]any{"bool":map[string]any{"should":[]any{map[string]any{"prefix":map[string]any{"e164":q}},map[string]any{"match_phrase_prefix":map[string]any{"display_name":q}}},"minimum_should_match":1}}},"filter":[]any{}}},"sort":[]any{map[string]any{"search_count":map[string]any{"order":"desc"}},map[string]any{"_score":map[string]any{"order":"desc"}}}}
 b:=query["query"].(map[string]any)["bool"].(map[string]any);if country!=""{b["filter"]=[]any{map[string]any{"term":map[string]any{"country_code":strings.ToUpper(country)}}}}
 raw,_:=json.Marshal(query);req,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.baseURL+"/"+PhoneIndex+"/_search",strings.NewReader(string(raw)));if err!=nil{return nil,err};req.Header.Set("Content-Type","application/json");resp,err:=c.http.Do(req);if err!=nil{return nil,err};defer resp.Body.Close();body,err:=io.ReadAll(io.LimitReader(resp.Body,2<<20));if err!=nil{return nil,err};if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("opensearch search returned %s",resp.Status)}
 var result struct{Hits struct{Hits []struct{Source PhoneSearchHit `json:"_source"`} `json:"hits"`} `json:"hits"`};if err:=json.Unmarshal(body,&result);err!=nil{return nil,err};out:=make([]PhoneSearchHit,0,len(result.Hits.Hits));for _,h:=range result.Hits.Hits{out=append(out,h.Source)};return out,nil
}
