package phone
import("context";"errors";"strings")
var ErrInvalidEvidence=errors.New("invalid verification evidence")
type EvidenceInput struct{Type string `json:"type"`;Value string `json:"value"`}
func(s Service)AddClaimEvidence(ctx context.Context,claimID string,in EvidenceInput)error{in.Type=strings.ToLower(strings.TrimSpace(in.Type));in.Value=strings.TrimSpace(in.Value);switch in.Type{case"website","document","dns","phone","other":default:return ErrInvalidEvidence};if in.Value==""||len(in.Value)>4000{return ErrInvalidEvidence};_,err:=s.Repository.DB.ExecContext(ctx,"INSERT INTO phone_verification_evidence(claim_id,evidence_type,evidence_value) VALUES($1,$2,$3)",claimID,in.Type,in.Value);return err}
