package phone
import "math"
func TrustScore(n Number,ids []Identity)float64{
 score:=0.0
 if n.VerificationStatus=="verified"{score+=45}
 best:=0.0;for _,i:=range ids{if i.ConfidenceScore>best{best=i.ConfidenceScore}}
 score+=best*.35
 score+=math.Min(n.DataQualityScore,100)*.20
 score-=n.SpamScore*35
 if n.VerificationStatus=="disputed"{score=math.Min(score,25)}
 if score<0{return 0};if score>100{return 100};return math.Round(score*100)/100
}
