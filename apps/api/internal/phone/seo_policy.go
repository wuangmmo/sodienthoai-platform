package phone
const (
 SitemapPageSize int64 = 40000
 MinIndexQuality float64 = 60
 MinIdentityConfidence float64 = 70
)
func CanonicalPath(e164 string) string { return "/phone/"+e164 }
func RobotsDirective(n Number,ids []Identity) string {
 if SEOStatusFor(n,ids)=="indexable"||n.SEOStatus=="indexed"{return "index,follow"}
 return "noindex,follow"
}
