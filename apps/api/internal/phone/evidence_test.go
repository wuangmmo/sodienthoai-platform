package phone
import "testing"
func TestEvidenceTypes(t *testing.T){valid:=[]string{"website","document","dns","phone","other"};for _,v:=range valid{if v==""{t.Fatal("empty evidence type")}}}
