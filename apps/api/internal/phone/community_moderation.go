package phone

import ("context";"errors";"strings")
var ErrInvalidCommentModeration=errors.New("invalid comment moderation")
func(s Service) ModerateComment(ctx context.Context,id,status string)error{
 status=strings.ToLower(strings.TrimSpace(status));if status!="approved"&&status!="rejected"&&status!="removed"{return ErrInvalidCommentModeration}
 res,err:=s.Repository.DB.ExecContext(ctx,`UPDATE phone_comments SET status=$2,reviewed_at=NOW() WHERE id=$1`,id,status);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrInvalidCommentModeration};return nil
}
