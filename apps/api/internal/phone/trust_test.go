package phone
import "testing"
func TestTrustScoreVerifiedIdentity(t *testing.T){n:=Number{VerificationStatus:"verified",DataQualityScore:90};s:=TrustScore(n,[]Identity{{ConfidenceScore:90}});if s<90{t.Fatalf("expected high trust, got %v",s)}}
func TestTrustScoreSpamPenalty(t *testing.T){n:=Number{VerificationStatus:"verified",DataQualityScore:80,SpamScore:1};s:=TrustScore(n,[]Identity{{ConfidenceScore:80}});if s>=70{t.Fatalf("spam should reduce trust, got %v",s)}}
func TestTrustScoreDisputedCap(t *testing.T){n:=Number{VerificationStatus:"disputed",DataQualityScore:100};s:=TrustScore(n,[]Identity{{ConfidenceScore:100}});if s>25{t.Fatalf("disputed trust must be capped, got %v",s)}}
