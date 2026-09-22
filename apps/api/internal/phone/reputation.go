package phone

import "math"

func SpamScore(approvedReports int64) float64 {
 if approvedReports<=0{return 0}
 score:=1-math.Exp(-float64(approvedReports)/5)
 if score>1{return 1};return math.Round(score*10000)/10000
}

func SEOStatusFor(n Number, identities []Identity) string {
 if n.VerificationStatus=="disputed"{return "review"}
 if len(identities)==0{return "noindex"}
 best:=float64(0);for _,i:=range identities{if i.ConfidenceScore>best{best=i.ConfidenceScore}}
 if n.VerificationStatus=="verified"&&best>=70&&n.DataQualityScore>=60{return "indexable"}
 return "noindex"
}
