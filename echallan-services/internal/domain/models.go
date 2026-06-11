package domain

import (
	"strings"
	"time"
)

type CustomDate struct {
	time.Time
}

func (c *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		c.Time = time.Time{}
		return nil
	}
	t, err := time.Parse("02-01-2006 15:04:05", s)
	if err != nil {
		return err
	}
	c.Time = t
	return nil
}

func (c CustomDate) MarshalJSON() ([]byte, error) {
	if c.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + c.Time.Format("02-01-2006 15:04:05") + `"`), nil
}

// RequestInfo handles the security and routing metadata
type RequestInfo struct {
	ApiId     string    `json:"apiId"`
	Ver       string    `json:"ver"`
	Ts        int64     `json:"ts"`
	Action    string    `json:"action"`
	Did       string    `json:"did,omitempty"`
	Key       string    `json:"key,omitempty"`
	MsgId     string    `json:"msgId"`
	AuthToken string    `json:"authToken,omitempty"`
	UserInfo  *UserInfo `json:"userInfo,omitempty"`
}

type UserInfo struct {
	MobileNumber      string      `json:"mobileNumber,omitempty"`
	EmailId           string      `json:"emailId,omitempty"`
	TenantId          string      `json:"tenantId"`
	Uuid              string      `json:"uuid"`
	UserName          string      `json:"userName"`
	Type              string      `json:"type"`
	Id                int         `json:"id,omitempty"`
	Roles             []Role      `json:"roles"`
	PwdExpiryDate     *CustomDate `json:"pwdExpiryDate,omitempty"`
	CreatedDate       *CustomDate `json:"createdDate,omitempty"`
	LastModifiedDate  *CustomDate `json:"lastModifiedDate,omitempty"`
}

type Role struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	TenantId string `json:"tenantId"`
}

type ResponseInfo struct {
	ApiId    string `json:"apiId"`
	Ver      string `json:"ver"`
	Ts       int64  `json:"ts"`
	ResMsgId string `json:"resMsgId,omitempty"`
	MsgId    string `json:"msgId"`
	Status   string `json:"status"`
}

type Error struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Params      string `json:"params,omitempty"`
}

type RequestInfoWrapper struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
}

type ChallanRequest struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
	Challan     *Challan     `json:"challan"`
}

type ChallanResponse struct {
	ResponseInfo         *ResponseInfo `json:"ResponseInfo"`
	Challans             []*Challan    `json:"challans"`
	CountOfServices      int           `json:"countOfServices,omitempty"`
	TotalAmountCollected int           `json:"totalAmountCollected,omitempty"`
	Validity             int           `json:"validity,omitempty"`
	TotalCount           int           `json:"totalCount,omitempty"`
}

type Challan struct {
	Id                  string          `json:"id,omitempty" db:"id"`
	TenantId            string          `json:"tenantId" db:"tenantid"`
	BusinessService     string          `json:"businessService" db:"businessservice"`
	ChallanNo           string          `json:"challanNo,omitempty" db:"challanno"`
	ReferenceId         string          `json:"referenceId,omitempty" db:"referenceid"`
	Description         string          `json:"description,omitempty" db:"description"`
	AccountId           string          `json:"accountId,omitempty" db:"accountid"`
	AdditionalDetail    interface{}     `json:"additionalDetail,omitempty"`
	Source              string          `json:"source,omitempty" db:"source"`
	TaxPeriodFrom       int64           `json:"taxPeriodFrom" db:"taxperiodfrom"`
	TaxPeriodTo         int64           `json:"taxPeriodTo" db:"taxperiodto"`
	Calculation         interface{}     `json:"calculation,omitempty"`
	Amount              []Amount        `json:"amount"`
	Address             *Address        `json:"address" db:"-"`
	Citizen             *UserInfo       `json:"citizen" db:"-"`
	ApplicationStatus   string          `json:"applicationStatus,omitempty" db:"applicationstatus"`
	Filestoreid         string          `json:"filestoreid,omitempty" db:"filestoreid"`
	ReceiptNumber       string          `json:"receiptNumber,omitempty" db:"receiptnumber"`
	AuditDetails        *AuditDetails   `json:"auditDetails,omitempty" db:"-"`
}

type Amount struct {
	TaxHeadCode string  `json:"taxHeadCode"`
	Amount      float64 `json:"amount"`
}

type Address struct {
	TenantId      string       `json:"tenantId"`
	DoorNo        string       `json:"doorNo,omitempty"`
	PlotNo        string       `json:"plotNo,omitempty"`
	Id            string       `json:"id,omitempty"`
	Landmark      string       `json:"landmark,omitempty"`
	City          string       `json:"city,omitempty"`
	Pincode       string       `json:"pincode,omitempty"`
	Detail        string       `json:"detail,omitempty"`
	BuildingName  string       `json:"buildingName,omitempty"`
	Street        string       `json:"street,omitempty"`
	Locality      *Boundary    `json:"locality,omitempty"`
	GeoLocation   *GeoLocation `json:"geoLocation,omitempty"`
}

type Boundary struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type GeoLocation struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type AuditDetails struct {
	CreatedBy        string `json:"createdBy"`
	LastModifiedBy   string `json:"lastModifiedBy"`
	CreatedTime      int64  `json:"createdTime"`
	LastModifiedTime int64  `json:"lastModifiedTime"`
}

type SearchCriteria struct {
	TenantId        string `form:"tenantId"`
	Ids             string `form:"ids"`
	ChallanNo       string `form:"challanNo"`
	AccountId       string `form:"accountId"`
	MobileNumber    string `form:"mobileNumber"`
	BusinessService string `form:"businessService"`
	Status          string `form:"status"`
	ReceiptNumber   string `form:"receiptNumber"`
	Offset          int    `form:"offset"`
	Limit           int    `form:"limit"`
}

// --- User Service DTOs ---
type CreateUserRequest struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
	User        *User        `json:"user"`
}
type UserDetailResponse struct {
	ResponseInfo *ResponseInfo `json:"responseInfo"`
	User         []*User       `json:"user"`
}

// --- IdGen Service DTOs ---
type IdGenerationRequest struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
	IdRequests  []IdRequest  `json:"idRequests"`
}
type IdRequest struct {
	IdName   string `json:"idName"`
	TenantId string `json:"tenantId"`
	Format   string `json:"format"`
}
type IdResponse struct {
	ResponseInfo *ResponseInfo `json:"responseInfo"`
	IdResponses  []struct {
		Id string `json:"id"`
	} `json:"idResponses"`
}

// --- Billing Service DTOs ---
type PaymentRequest struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
	Payment     *Payment     `json:"Payment"`
}
type Payment struct {
	TenantId      string `json:"tenantId"`
	PaymentMode   string `json:"paymentMode"`
	PaidBy        string `json:"paidBy"`
	TotalAmount   int    `json:"totalAmountPaid"`
	ReceiptNumber string `json:"receiptNumber"`
}
type BillDetail struct {
	BillId string `json:"billId"`
}

type CalculationCriteria struct {
	TenantId  string   `json:"tenantId"`
	ChallanNo string   `json:"challanNo"`
	Challan   *Challan `json:"challan"`
}

type CalculationReq struct {
	RequestInfo         *RequestInfo          `json:"RequestInfo"`
	CalculationCriteria []CalculationCriteria `json:"CalulationCriteria"`
}

// --- Missing User Definition ---
type User struct {
	Id           int64  `json:"id,omitempty"`
	UserName     string `json:"userName"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	MobileNumber string `json:"mobileNumber"`
	EmailId      string `json:"emailId,omitempty"`
	TenantId     string `json:"tenantId"`
	PwdExpiryDate     *CustomDate `json:"pwdExpiryDate,omitempty"`
	CreatedDate       *CustomDate `json:"createdDate,omitempty"`
	LastModifiedDate  *CustomDate `json:"lastModifiedDate,omitempty"`
}
