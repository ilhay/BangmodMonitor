package handler

import (
	"net/http"
	"time"

	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	maria     *storage.Maria
	jwtSecret string
	stripe    interface {
		Enabled() bool
		CreateCustomer(email, orgName string) (string, error)
	}
}

func NewAuth(maria *storage.Maria, jwtSecret string, stripe interface {
	Enabled() bool
	CreateCustomer(email, orgName string) (string, error)
}) *AuthHandler {
	return &AuthHandler{maria: maria, jwtSecret: jwtSecret, stripe: stripe}
}

type registerReq struct {
	OrgName  string `json:"org_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check email not taken
	existing, err := h.maria.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}

	orgID := uuid.New().String()
	userID := uuid.New().String()

	if err := h.maria.CreateOrg(c.Request.Context(), orgID, req.OrgName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create org failed"})
		return
	}
	if err := h.maria.CreateUser(c.Request.Context(), userID, orgID, req.Email, string(hash), "admin"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
		return
	}

	// Assign free plan subscription
	subID := uuid.New().String()
	_ = h.maria.CreateSubscription(c.Request.Context(), subID, orgID, "plan-free")

	// Auto-enable 1 default region (th)
	_ = h.maria.AddOrgRegion(c.Request.Context(), uuid.New().String(), orgID, "th")

	// Create Stripe customer if configured
	if h.stripe.Enabled() {
		if stripeCustomerID, err := h.stripe.CreateCustomer(req.Email, req.OrgName); err == nil {
			_ = h.maria.UpdateSubscriptionStripe(c.Request.Context(), orgID, stripeCustomerID, "", "active", nil, false)
		}
	}

	token, err := h.issueJWT(userID, orgID, req.Email, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":   token,
		"user_id": userID,
		"org_id":  orgID,
		"email":   req.Email,
	})
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.maria.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.issueJWT(user.ID, user.OrgID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user_id": user.ID,
		"org_id":  user.OrgID,
		"email":   user.Email,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id": c.GetString("user_id"),
		"org_id":  c.GetString("org_id"),
		"email":   c.GetString("email"),
		"role":    c.GetString("role"),
	})
}

func (h *AuthHandler) issueJWT(userID, orgID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"org_id":  orgID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
}

