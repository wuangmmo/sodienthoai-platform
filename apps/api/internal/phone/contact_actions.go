package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidContactAction=errors.New("invalid contact action")
var contactActions=map[string]bool{"keep_local":true,"update_local":true,"suggest_public_correction":true,"report_inaccurate":true,"invite_owner":true}

func(s Service) ContactAction(ctx context.Context,ownerKey,contactID,action string)(string,error){
 action=strings.ToLower(strings.TrimSpace(action));if !contactActions[action]{return "",ErrInvalidContactAction}
 var id string
 err:=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO contact_identity_actions(contact_id,action_type)
SELECT c.id,$3 FROM private_contacts c JOIN contact_books b ON b.id=c.contact_book_id WHERE c.id=$2 AND b.owner_key=$1 RETURNING id::text`,ownerKey,contactID,action).Scan(&id)
 return id,err
}

func(s Service) DeleteContactBook(ctx context.Context,ownerKey string)error{
 _,err:=s.Repository.DB.ExecContext(ctx,`DELETE FROM contact_books WHERE owner_key=$1`,ownerKey);return err
}
