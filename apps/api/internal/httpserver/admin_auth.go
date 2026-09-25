package httpserver

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AdminAuthenticator struct {
	DB       *sql.DB
	Secret   string
	Issuer   string
	Audience string
	TTL      time.Duration
}

type adminLoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type adminDTO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
}

type adminClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Iss   string `json:"iss"`
	Aud   string `json:"aud"`
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
}

func (a AdminAuthenticator) Login(w http.ResponseWriter, r *http.Request) {
	var in adminLoginInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&in) != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid_request"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" || len(in.Password) > 1024 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error":"invalid_credentials"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var admin adminDTO
	var passwordHash string
	err := a.DB.QueryRowContext(ctx, `SELECT id::text,name,email,password_hash,role,status,created_at,last_login_at
		FROM admin_users WHERE LOWER(email)=LOWER($1) LIMIT 1`, email).
		Scan(&admin.ID,&admin.Name,&admin.Email,&passwordHash,&admin.Role,&admin.Status,&admin.CreatedAt,&admin.LastLoginAt)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.Password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error":"invalid_credentials"})
		return
	}
	if admin.Status == "disabled" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error":"admin_disabled"})
		return
	}
	now := time.Now().UTC()
	if _, err := a.DB.ExecContext(ctx, `UPDATE admin_users SET last_login_at=$1,updated_at=$1 WHERE id=$2::uuid`, now, admin.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error":"internal_error"})
		return
	}
	admin.LastLoginAt = &now
	token, err := a.issue(admin, now)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error":"internal_error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"admin":admin,"token":token})
}

func (a AdminAuthenticator) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.authenticate(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error":"unauthorized"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var admin adminDTO
	err := a.DB.QueryRowContext(ctx, `SELECT id::text,name,email,role,status,created_at,last_login_at
		FROM admin_users WHERE id=$1::uuid LIMIT 1`, claims.Sub).
		Scan(&admin.ID,&admin.Name,&admin.Email,&admin.Role,&admin.Status,&admin.CreatedAt,&admin.LastLoginAt)
	if err != nil || admin.Status == "disabled" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error":"unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"admin":admin})
}

func (a AdminAuthenticator) Logout(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.authenticate(r); !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error":"unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok":true})
}

func (a AdminAuthenticator) issue(admin adminDTO, now time.Time) (string,error) {
	if len(a.Secret) < 32 { return "", errors.New("admin session secret is not configured") }
	ttl := a.TTL
	if ttl <= 0 { ttl = 8*time.Hour }
	header, _ := json.Marshal(map[string]string{"alg":"HS256","typ":"JWT"})
	payload, _ := json.Marshal(adminClaims{Sub:admin.ID,Email:admin.Email,Role:admin.Role,Iss:a.Issuer,Aud:a.Audience,Iat:now.Unix(),Exp:now.Add(ttl).Unix()})
	unsigned := base64.RawURLEncoding.EncodeToString(header)+"."+base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(a.Secret)); _, _ = mac.Write([]byte(unsigned))
	return unsigned+"."+base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (a AdminAuthenticator) authenticate(r *http.Request) (adminClaims,bool) {
	var zero adminClaims
	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(raw, "Bearer ") || len(a.Secret) < 32 { return zero,false }
	token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
	parts := strings.Split(token, ".")
	if len(parts) != 3 { return zero,false }
	unsigned := parts[0]+"."+parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2]); if err != nil { return zero,false }
	mac := hmac.New(sha256.New, []byte(a.Secret)); _, _ = mac.Write([]byte(unsigned))
	if !hmac.Equal(sig, mac.Sum(nil)) { return zero,false }
	body, err := base64.RawURLEncoding.DecodeString(parts[1]); if err != nil { return zero,false }
	var claims adminClaims
	if json.Unmarshal(body,&claims) != nil { return zero,false }
	now := time.Now().Unix()
	if claims.Sub=="" || claims.Exp<=now || claims.Iss!=a.Issuer || claims.Aud!=a.Audience { return zero,false }
	if claims.Role!="admin" && claims.Role!="super-admin" { return zero,false }
	return claims,true
}
