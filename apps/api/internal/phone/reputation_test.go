package phone

import "testing"

func TestSpamScore(t *testing.T){
 if SpamScore(0)!=0{t.Fatal("zero reports must be zero")}
 if SpamScore(10)<=SpamScore(1){t.Fatal("score must increase")}
 if SpamScore(100)>1{t.Fatal("score must not exceed one")}
}
func TestSEOStatusFor(t *testing.T){
 n:=Number{VerificationStatus:"verified",DataQualityScore:80}
 if SEOStatusFor(n,nil)!="noindex"{t.Fatal("identity required")}
 ids:=[]Identity{{ConfidenceScore:80}}
 if SEOStatusFor(n,ids)!="indexable"{t.Fatal("verified quality identity should be indexable")}
 n.VerificationStatus="disputed";if SEOStatusFor(n,ids)!="review"{t.Fatal("disputed must be review")}
}
