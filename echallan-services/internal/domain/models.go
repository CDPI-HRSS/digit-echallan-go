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
	Name              string      `json:"name,omitempty"`
	UserName          string      `json:"userName"`
	Type              string      `json:"type"`
	Id                int64       `json:"id,omitempty"`
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
	RequestInfo *RequestInfo `json:"RequestInfo" binding:"required"`
	Challan     *Challan     `json:"challan" binding:"required"`
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
	Id                  string          `json:"id,omitempty" db:"id" gorm:"column:id;primaryKey"`
	TenantId            string          `json:"tenantId" db:"tenantid" gorm:"column:tenantid" binding:"required"`
	BusinessService     string          `json:"businessService" db:"businessservice" gorm:"column:businessservice" binding:"required"`
	ChallanNo           string          `json:"challanNo,omitempty" db:"challanno" gorm:"column:challanno"`
	ReferenceId         string          `json:"referenceId,omitempty" db:"referenceid" gorm:"column:referenceid"`
	Description         string          `json:"description,omitempty" db:"description" gorm:"column:description"`
	AccountId           string          `json:"accountId,omitempty" db:"accountid" gorm:"column:accountid"`
	AdditionalDetail    interface{}     `json:"additionalDetail,omitempty" gorm:"-"`
	Source              string          `json:"source,omitempty" db:"source" gorm:"column:source"`
	TaxPeriodFrom       int64           `json:"taxPeriodFrom" db:"taxperiodfrom" gorm:"column:taxperiodfrom" binding:"required,gt=0"`
	TaxPeriodTo         int64           `json:"taxPeriodTo" db:"taxperiodto" gorm:"column:taxperiodto" binding:"required,gt=0"`
	Calculation         interface{}     `json:"calculation,omitempty" gorm:"-"`
	Amount              []Amount        `json:"amount" gorm:"-"`
	Address             *Address        `json:"address" db:"-" gorm:"-" binding:"required"`
	Citizen             *UserInfo       `json:"citizen" db:"-" gorm:"-" binding:"required"`
	ApplicationStatus   string          `json:"applicationStatus,omitempty" db:"applicationstatus" gorm:"column:applicationstatus"`
	Filestoreid         string          `json:"filestoreid,omitempty" db:"filestoreid" gorm:"column:filestoreid"`
	ReceiptNumber       string          `json:"receiptNumber,omitempty" db:"receiptnumber" gorm:"column:receiptnumber"`
	AuditDetails        *AuditDetails   `json:"auditDetails,omitempty" db:"-" gorm:"-"`
}

type ChallanDB struct {
	Id                string             `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantId          string             `json:"tenantId" gorm:"column:tenantid"`
	BusinessService   string             `json:"businessService" gorm:"column:businessservice"`
	ChallanNo         string             `json:"challanNo,omitempty" gorm:"column:challanno"`
	ReferenceId       string             `json:"referenceId,omitempty" gorm:"column:referenceid"`
	Description       string             `json:"description,omitempty" gorm:"column:description"`
	AccountId         string             `json:"accountId,omitempty" gorm:"column:accountid"`
	Source            string             `json:"source,omitempty" gorm:"column:source"`
	TaxPeriodFrom     int64              `json:"taxPeriodFrom" gorm:"column:taxperiodfrom"`
	TaxPeriodTo       int64              `json:"taxPeriodTo" gorm:"column:taxperiodto"`
	ApplicationStatus string             `json:"applicationStatus,omitempty" gorm:"column:applicationstatus"`
	Filestoreid       string             `json:"filestoreid,omitempty" gorm:"column:filestoreid"`
	ReceiptNumber     string             `json:"receiptNumber,omitempty" gorm:"column:receiptnumber"`
	CreatedBy         string             `gorm:"column:createdby"`
	LastModifiedBy    string             `gorm:"column:lastmodifiedby"`
	CreatedTime       int64              `gorm:"column:createdtime"`
	LastModifiedTime  int64              `gorm:"column:lastmodifiedtime"`
	Amounts           []ChallanAmountDB  `gorm:"foreignKey:EchallanId;references:Id"`
	AddressDB         *ChallanAddressDB  `gorm:"foreignKey:EchallanId;references:Id"`
}

func (ChallanDB) TableName() string { return "eg_echallan" }

type ChallanAmountDB struct {
	Id          string  `gorm:"column:id;primaryKey"`
	EchallanId  string  `gorm:"column:echallanid"`
	TaxHeadCode string  `gorm:"column:taxheadcode"`
	Amount      float64 `gorm:"column:amount"`
}

func (ChallanAmountDB) TableName() string { return "eg_challan_amount" }

type ChallanAddressDB struct {
	Id           string  `gorm:"column:id;primaryKey"`
	EchallanId   string  `gorm:"column:echallanid"`
	DoorNo       string  `gorm:"column:doorno"`
	BuildingName string  `gorm:"column:buildingname"`
	Street       string  `gorm:"column:street"`
	City         string  `gorm:"column:city"`
	Pincode      string  `gorm:"column:pincode"`
	Latitude     float64 `gorm:"column:latitude"`
	Longitude    float64 `gorm:"column:longitude"`
	LocalityCode string  `gorm:"column:locality_code"`
}

func (ChallanAddressDB) TableName() string { return "eg_challan_address" }

func (db *ChallanDB) ToChallan() *Challan {
	c := &Challan{
		Id:                db.Id,
		TenantId:          db.TenantId,
		BusinessService:   db.BusinessService,
		ChallanNo:         db.ChallanNo,
		ReferenceId:       db.ReferenceId,
		Description:       db.Description,
		AccountId:         db.AccountId,
		Source:            db.Source,
		TaxPeriodFrom:     db.TaxPeriodFrom,
		TaxPeriodTo:       db.TaxPeriodTo,
		ApplicationStatus: db.ApplicationStatus,
		Filestoreid:       db.Filestoreid,
		ReceiptNumber:     db.ReceiptNumber,
		AuditDetails: &AuditDetails{
			CreatedBy:        db.CreatedBy,
			LastModifiedBy:   db.LastModifiedBy,
			CreatedTime:      db.CreatedTime,
			LastModifiedTime: db.LastModifiedTime,
		},
	}
	for _, amt := range db.Amounts {
		c.Amount = append(c.Amount, Amount{
			TaxHeadCode: amt.TaxHeadCode,
			Amount:      amt.Amount,
		})
	}
	if db.AddressDB != nil {
		c.Address = &Address{
			DoorNo:       db.AddressDB.DoorNo,
			BuildingName: db.AddressDB.BuildingName,
			Street:       db.AddressDB.Street,
			City:         db.AddressDB.City,
			Pincode:      db.AddressDB.Pincode,
			Locality: &Boundary{
				Code: db.AddressDB.LocalityCode,
			},
			GeoLocation: &GeoLocation{
				Latitude:  db.AddressDB.Latitude,
				Longitude: db.AddressDB.Longitude,
			},
		}
	}
	return c
}

func (c *Challan) ToChallanDB() ChallanDB {
	db := ChallanDB{
		Id:                c.Id,
		TenantId:          c.TenantId,
		BusinessService:   c.BusinessService,
		ChallanNo:         c.ChallanNo,
		ReferenceId:       c.ReferenceId,
		Description:       c.Description,
		AccountId:         c.AccountId,
		Source:            c.Source,
		TaxPeriodFrom:     c.TaxPeriodFrom,
		TaxPeriodTo:       c.TaxPeriodTo,
		ApplicationStatus: c.ApplicationStatus,
		Filestoreid:       c.Filestoreid,
		ReceiptNumber:     c.ReceiptNumber,
	}
	if c.AuditDetails != nil {
		db.CreatedBy = c.AuditDetails.CreatedBy
		db.LastModifiedBy = c.AuditDetails.LastModifiedBy
		db.CreatedTime = c.AuditDetails.CreatedTime
		db.LastModifiedTime = c.AuditDetails.LastModifiedTime
	}
	for _, amt := range c.Amount {
		db.Amounts = append(db.Amounts, ChallanAmountDB{
			EchallanId:  c.Id,
			TaxHeadCode: amt.TaxHeadCode,
			Amount:      amt.Amount,
		})
	}
	if c.Address != nil {
		db.AddressDB = &ChallanAddressDB{
			EchallanId:   c.Id,
			DoorNo:       c.Address.DoorNo,
			BuildingName: c.Address.BuildingName,
			Street:       c.Address.Street,
			City:         c.Address.City,
			Pincode:      c.Address.Pincode,
		}
		if c.Address.Locality != nil {
			db.AddressDB.LocalityCode = c.Address.Locality.Code
		}
		if c.Address.GeoLocation != nil {
			db.AddressDB.Latitude = c.Address.GeoLocation.Latitude
			db.AddressDB.Longitude = c.Address.GeoLocation.Longitude
		}
	}
	return db
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
	CalculationCriteria []CalculationCriteria `json:"CalculationCriteria"`
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
