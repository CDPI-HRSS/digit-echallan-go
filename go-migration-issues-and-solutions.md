# Go Migration Issues and Recommended Solutions

This document summarizes the main issues found during the evaluation of the Go-based eChallan migration for both the eChallan services and calculator services, along with practical solutions.

## 1. REST API design

### Problem

The current endpoints use action-style paths such as:

- `/_create`
- `/_search`
- `/_update`
- `/_count`
- `/_calculate`

These are not fully aligned with idiomatic RESTful Go/API design.

### Impact

- Less intuitive for consumers
- Harder to maintain and evolve
- Not aligned with the migration guidance that prefers noun-based resource routes

### Solution

- Move to resource-oriented routes such as:
  - `POST /challans`
  - `GET /challans/:id`
  - `POST /challans/search`
  - `PUT /challans/:id`
  - `POST /challans/count`
- Keep legacy routes temporarily as backward-compatible aliases if required.
- Use consistent versioning such as `/api/v1/challans`.

Example:

Bad:

```go
func (cc *ChallanController) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/eChallan/v1")
	{
		v1.POST("/_create", cc.Create)
		v1.POST("/_search", cc.Search)
	}
}
```

Good:

```go
func (cc *ChallanController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/challans")
	{
		api.POST("", cc.Create)
		api.GET(":id", cc.GetByID)
		api.POST("/search", cc.Search)
	}
}
```

---

## 2. Missing authentication and authorization

### Problem

The current Go code does not show implementation of Keycloak/JWT validation or authorization middleware.

### Impact

- Security controls from the Java version may not be preserved
- Protected routes may be exposed without proper access checks

### Solution

- Add JWT/OIDC validation middleware.
- Integrate Keycloak token introspection or validation flow.
- Protect sensitive routes using middleware.
- Map roles and permissions to Go middleware or service checks.

Example:

Bad:

```go
func (cc *ChallanController) Create(c *gin.Context) {
	// No authentication check
	c.Next()
}
```

Good:

```go
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
```

```go
r.Use(AuthMiddleware())
```

---

## 3. ORM strategy mismatch

### Problem

The service uses SQLX and raw query building instead of GORM, while the migration checklist expected GORM-style model mapping and ORM conventions.

### Impact

- The code is valid Go, but it does not fully follow the intended migration pattern for the Java-to-Go transition.
- Database model mapping may be less consistent and less maintainable.

### Solution

- Either:
  - continue with SQLX but formalize repository and query patterns, or
  - migrate to GORM for model mapping and associations where suitable.
- If staying with SQLX, ensure strong query abstractions, prepared statements, and clear model tags.

Example:

Bad:

```go
rows, err := db.Queryx("SELECT id, tenant_id FROM challans WHERE tenant_id = $1", tenantID)
```

Good:

```go
type Challan struct {
	ID       string `gorm:"primaryKey"`
	TenantID string `gorm:"column:tenant_id"`
}

_ = db.AutoMigrate(&Challan{})
```

---

## 4. Testing is present but still lightweight

### Problem

The repository includes tests, but they are mostly structural and not comprehensive enough for strong migration confidence.

### Impact

- Lower confidence in business logic correctness
- Limited protection against regressions

### Solution

- Add table-driven unit tests for service behavior.
- Add controller tests for request validation and response handling.
- Add integration tests for repository and Kafka interactions.
- Increase coverage for edge cases and error paths.

Example:

Bad:

```go
func TestCreateChallan(t *testing.T) {
	// only one simple smoke test
	if false {
		t.Error("not implemented")
	}
}
```

Good:

```go
func TestCreateChallan(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    int
	}{
		{name: "invalid json", payload: "{invalid}", want: http.StatusBadRequest},
		{name: "valid json", payload: `{"tenantId":"pb.amritsar"}`, want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange router / handler
			// assert status code
		})
	}
}
```

---

## 5. Error handling consistency

### Problem

Some handlers return errors directly while others use custom response structures. The response style is not fully uniform.

### Impact

- Inconsistent client-facing error responses
- Harder debugging and maintenance

### Solution

- Standardize error responses across controllers.
- Use consistent error codes and messages.
- Keep business errors in the service layer and map them to HTTP responses in the transport layer.

Example:

Bad:

```go
if err != nil {
	c.JSON(400, gin.H{"error": err.Error()})
}
```

Good:

```go
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

if err != nil {
	c.JSON(http.StatusBadRequest, APIError{
		Code:    "VALIDATION_ERROR",
		Message: err.Error(),
	})
}
```

---

## 6. Validation and request handling

### Problem

Validation is present in some places, but the handlers and services could be more consistent and explicit.

### Impact

- Some invalid requests may be handled less cleanly
- Response quality may vary by endpoint

### Solution

- Centralize validation in dedicated validator packages.
- Return structured validation errors consistently.
- Use strong binding and validation rules for request payloads.

Example:

Bad:

```go
func (cc *ChallanController) Create(c *gin.Context) {
	var req map[string]interface{}
	_ = c.ShouldBindJSON(&req)
	// no validation
}
```

Good:

```go
type CreateChallanRequest struct {
	TenantID        string `json:"tenantId" binding:"required"`
	BusinessService string `json:"businessService" binding:"required"`
}

func (cc *ChallanController) Create(c *gin.Context) {
	var req CreateChallanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
}
```

---

## 7. Service layering and dependency injection

### Problem

The code already uses a layered structure, but some constructors and dependencies are still tightly coupled to concrete implementations.

### Impact

- Harder to mock and test
- More fragile when dependencies change

### Solution

- Keep the current layered approach.
- Prefer interfaces at service boundaries.
- Inject dependencies through constructors for better testability.

Example:

Bad:

```go
type ChallanController struct {
	service *ChallanServiceImpl
}
```

Good:

```go
type ChallanService interface {
	Create(req *ChallanRequest) (*Challan, error)
}
```

```go
type ChallanController struct {
	service ChallanService
}

func NewChallanController(service ChallanService) *ChallanController {
	return &ChallanController{service: service}
}
```

---

## 8. Documentation and operational readiness

### Problem

The Go services are functional, but some deployment and operational concerns are still not fully documented for production readiness.

### Impact

- Harder onboarding and troubleshooting
- Increased operational risk

### Solution

- Add deployment notes for Docker/Kubernetes.
- Document environment variables, Kafka topics, DB settings, and auth integration.
- Include runbooks for health checks and common failures.

Example:

Bad:

```go
// hardcoded values inside code
var port = "8082"
```

Good:

```env
SERVER_PORT=8082
DB_HOST=localhost
DB_PORT=5432
KAFKA_BROKERS=localhost:9092
```