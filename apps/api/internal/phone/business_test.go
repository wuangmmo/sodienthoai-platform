package phone

import "testing"

func TestValidSlug(t *testing.T){
 good:=[]string{"cong-ty-a","taxi-lam-ha","abc123"}
 bad:=[]string{"A","-abc","abc-","abc--def","abc def","điện-thoại"}
 for _,v:=range good{if !validSlug(v){t.Fatalf("expected valid slug %q",v)}}
 for _,v:=range bad{if validSlug(v){t.Fatalf("expected invalid slug %q",v)}}
}
