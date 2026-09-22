package phone
import "testing"
func TestRobotsDirective(t *testing.T){n:=Number{VerificationStatus:"unverified",SEOStatus:"noindex"};if RobotsDirective(n,nil)!="noindex,follow"{t.Fatal("unknown number must be noindex")};n.VerificationStatus="verified";n.DataQualityScore=80;ids:=[]Identity{{ConfidenceScore:90}};if RobotsDirective(n,ids)!="index,follow"{t.Fatal("qualified number must be indexable")}}
func TestCanonicalPath(t *testing.T){if CanonicalPath("+84901234567")!="/phone/+84901234567"{t.Fatal("canonical path changed")}}
