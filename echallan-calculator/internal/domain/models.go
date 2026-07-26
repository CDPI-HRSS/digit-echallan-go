package domain

// RequestInfo handles the security and routing metadata
type RequestInfo struct {
	APIId       string    `json:"apiId,omitempty"`
	Ver         string    `json:"ver,omitempty"`
	Ts          int64     `json:"ts,omitempty"`
	Action      string    `json:"action,omitempty"`
	Did         string    `json:"did,omitempty"`
	Key         string    `json:"key,omitempty"`
	MsgId       string    `json:"msgId,omitempty"`
	RequesterId string    `json:"requesterId,omitempty"`
	AuthToken   string    `json:"authToken,omitempty"`
	UserInfo    *UserInfo `json:"userInfo,omitempty"`
}

type User struct {
	Id                    int64  `json:"id,omitempty"`
	Uuid                  string `json:"uuid,omitempty"`
	UserName              string `json:"userName,omitempty"`
	Password              string `json:"password,omitempty"`
	Salutation            string `json:"salutation,omitempty"`
	Name                  string `json:"name,omitempty"`
	Gender                string `json:"gender,omitempty"`
	MobileNumber          string `json:"mobileNumber,omitempty"`
	EmailId               string `json:"emailId,omitempty"`
	AltContactNumber      string `json:"altContactNumber,omitempty"`
	Pan                   string `json:"pan,omitempty"`
	AadhaarNumber         string `json:"aadhaarNumber,omitempty"`
	PermanentAddress      string `json:"permanentAddress,omitempty"`
	PermanentCity         string `json:"permanentCity,omitempty"`
	PermanentPinCode      string `json:"permanentPinCode,omitempty"`
	CorrespondenceCity    string `json:"correspondenceCity,omitempty"`
	CorrespondencePinCode string `json:"correspondencePinCode,omitempty"`
	CorrespondenceAddress string `json:"correspondenceAddress,omitempty"`
	Active                *bool  `json:"active,omitempty"`
	Roles                 []Role `json:"roles,omitempty"`
	TenantId              string `json:"tenantId,omitempty"`
	Type                  string `json:"type,omitempty"`
}

type Role struct {
	Id       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Code     string `json:"code,omitempty"`
	TenantId string `json:"tenantId,omitempty"`
}

type TenantRole struct {
	TenantId string `json:"tenantId,omitempty"`
	Roles    []Role `json:"roles,omitempty"`
}

type UserInfo struct {
	User
	TenantId        string       `json:"tenantId,omitempty"`
	Uuid            string       `json:"uuid,omitempty"`
	UserName        string       `json:"userName,omitempty"`
	Password        string       `json:"password,omitempty"`
	IdToken         string       `json:"idToken,omitempty"`
	MobileNumber    string       `json:"mobileNumber,omitempty"`
	Email           string       `json:"email,omitempty"`
	PrimaryRole     []Role       `json:"primaryrole,omitempty"`
	AdditionalRoles []TenantRole `json:"additionalroles,omitempty"`
}

// Convert UserInfo/User helper to standard common user for payer
func (ui *UserInfo) ToCommonUser() *User {
	if ui == nil {
		return nil
	}
	u := User{
		Id:                    ui.Id,
		Uuid:                  ui.Uuid,
		UserName:              ui.UserName,
		Password:              ui.Password,
		Salutation:            ui.Salutation,
		Name:                  ui.Name,
		Gender:                ui.Gender,
		MobileNumber:          ui.MobileNumber,
		EmailId:               ui.EmailId,
		AltContactNumber:      ui.AltContactNumber,
		Pan:                   ui.Pan,
		AadhaarNumber:         ui.AadhaarNumber,
		PermanentAddress:      ui.PermanentAddress,
		PermanentCity:         ui.PermanentCity,
		PermanentPinCode:      ui.PermanentPinCode,
		CorrespondenceCity:    ui.CorrespondenceCity,
		CorrespondencePinCode: ui.CorrespondencePinCode,
		CorrespondenceAddress: ui.CorrespondenceAddress,
		Active:                ui.Active,
		TenantId:              ui.TenantId,
		Type:                  ui.Type,
	}

	if len(ui.User.Roles) > 0 {
		u.Roles = make([]Role, len(ui.User.Roles))
		copy(u.Roles, ui.User.Roles)
	}

	if u.TenantId == "" {
		u.TenantId = ui.TenantId
	}
	if u.Uuid == "" {
		u.Uuid = ui.Uuid
	}
	if u.UserName == "" {
		u.UserName = ui.UserName
	}
	if u.MobileNumber == "" {
		u.MobileNumber = ui.MobileNumber
	}
	return &u
}

type ResponseInfo struct {
	APIId    string `json:"apiId,omitempty"`
	Ver      string `json:"ver,omitempty"`
	Ts       int64  `json:"ts,omitempty"`
	ResMsgId string `json:"resMsgId,omitempty"`
	MsgId    string `json:"msgId,omitempty"`
	Status   string `json:"status,omitempty"`
}

type AuditDetails struct {
	CreatedBy        string `json:"createdBy,omitempty"`
	LastModifiedBy   string `json:"lastModifiedBy,omitempty"`
	CreatedTime      int64  `json:"createdTime,omitempty"`
	LastModifiedTime int64  `json:"lastModifiedTime,omitempty"`
}

type Amount struct {
	TaxHeadCode string  `json:"taxHeadCode"`
	Amount      float64 `json:"amount"`
}

type Boundary struct {
	Code             string     `json:"code"`
	Name             string     `json:"name,omitempty"`
	Label            string     `json:"label,omitempty"`
	Latitude         string     `json:"latitude,omitempty"`
	Longitude        string     `json:"longitude,omitempty"`
	Children         []Boundary `json:"children,omitempty"`
	MaterializedPath string     `json:"materializedPath,omitempty"`
}

type Address struct {
	Id             string    `json:"id,omitempty"`
	TenantId       string    `json:"tenantId,omitempty"`
	DoorNo         string    `json:"doorNo,omitempty"`
	Latitude       float64   `json:"latitude,omitempty"`
	Longitude      float64   `json:"longitude,omitempty"`
	AddressId      string    `json:"addressId,omitempty"`
	AddressNumber  string    `json:"addressNumber,omitempty"`
	Type           string    `json:"type,omitempty"`
	AddressLine1   string    `json:"addressLine1,omitempty"`
	AddressLine2   string    `json:"addressLine2,omitempty"`
	Landmark       string    `json:"landmark,omitempty"`
	City           string    `json:"city,omitempty"`
	Pincode        string    `json:"pincode,omitempty"`
	Detail         string    `json:"detail,omitempty"`
	BuildingName   string    `json:"buildingName,omitempty"`
	Street         string    `json:"street,omitempty"`
	Locality       *Boundary `json:"locality,omitempty"`
	PlotNo         string    `json:"plotNo,omitempty"`
	District       string    `json:"district,omitempty"`
	State          string    `json:"state,omitempty"`
	Country        string    `json:"country,omitempty"`
	Region         string    `json:"region,omitempty"`
}

type Challan struct {
	Citizen           *UserInfo     `json:"citizen,omitempty"`
	Id                string        `json:"id,omitempty"`
	TenantId          string        `json:"tenantId,omitempty"`
	BusinessService   string        `json:"businessService,omitempty"`
	ChallanNo         string        `json:"challanNo,omitempty"`
	ReferenceId       string        `json:"referenceId,omitempty"`
	Description       string        `json:"description,omitempty"`
	AccountId         string        `json:"accountId,omitempty"`
	AdditionalDetail  interface{}   `json:"additionalDetail,omitempty"`
	ApplicationStatus string        `json:"applicationStatus,omitempty"`
	Source            string        `json:"source,omitempty"`
	TaxPeriodFrom     int64         `json:"taxPeriodFrom,omitempty"`
	TaxPeriodTo       int64         `json:"taxPeriodTo,omitempty"`
	Calculation       *Calculation  `json:"calculation,omitempty"`
	Amount            []Amount      `json:"amount,omitempty"`
	Address           *Address      `json:"address,omitempty"`
	AuditDetails      *AuditDetails `json:"auditDetails,omitempty"`
}

type RequestInfoWrapper struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
}

type ChallanResponse struct {
	ResponseInfo *ResponseInfo `json:"ResponseInfo,omitempty"`
	Challans     []Challan     `json:"challans,omitempty"`
}

type CalculationCriteria struct {
	Challan   *Challan `json:"challan,omitempty"`
	ChallanNo string   `json:"challanNo,omitempty"`
	TenantId  string   `json:"tenantId" binding:"required"`
}

type TaxHeadEstimate struct {
	TaxHeadCode    string  `json:"taxHeadCode"`
	EstimateAmount float64 `json:"estimateAmount"`
	Category       string  `json:"category,omitempty"`
}

type Calculation struct {
	ChallanNo        string            `json:"challanNo,omitempty"`
	Challan          *Challan          `json:"challan,omitempty"`
	TenantId         string            `json:"tenantId"`
	TaxHeadEstimates []TaxHeadEstimate `json:"taxHeadEstimates"`
}

type CalculationReq struct {
	RequestInfo        *RequestInfo         `json:"RequestInfo" binding:"required"`
	CalculationCriteria []CalculationCriteria `json:"CalculationCriteria" binding:"required"`
}

type CalculationRes struct {
	ResponseInfo *ResponseInfo `json:"ResponseInfo,omitempty"`
	Calculations []Calculation `json:"Calculations"`
}

type DemandDetail struct {
	Id                string        `json:"id,omitempty"`
	DemandId          string        `json:"demandId,omitempty"`
	TaxHeadMasterCode string        `json:"taxHeadMasterCode"`
	TaxAmount         float64       `json:"taxAmount"`
	CollectionAmount  float64       `json:"collectionAmount"`
	AdditionalDetails interface{}   `json:"additionalDetails,omitempty"`
	AuditDetails      *AuditDetails `json:"auditDetails,omitempty"`
	TenantId          string        `json:"tenantId,omitempty"`
}

type Demand struct {
	Id                   string         `json:"id,omitempty"`
	TenantId             string         `json:"tenantId"`
	ConsumerCode         string         `json:"consumerCode"`
	ConsumerType         string         `json:"consumerType,omitempty"`
	BusinessService      string         `json:"businessService"`
	Payer                *User          `json:"payer,omitempty"`
	TaxPeriodFrom        int64          `json:"taxPeriodFrom"`
	TaxPeriodTo          int64          `json:"taxPeriodTo"`
	DemandDetails        []DemandDetail `json:"demandDetails"`
	AuditDetails         *AuditDetails  `json:"auditDetails,omitempty"`
	AdditionalDetails    interface{}    `json:"additionalDetails,omitempty"`
	MinimumAmountPayable float64        `json:"minimumAmountPayable"`
	Status               string         `json:"status,omitempty"`
}

type DemandRequest struct {
	RequestInfo *RequestInfo `json:"RequestInfo"`
	Demands     []Demand     `json:"Demands"`
}

type DemandResponse struct {
	ResponseInfo *ResponseInfo `json:"ResponseInfo,omitempty"`
	Demands      []Demand      `json:"Demands"`
}

type GenerateBillCriteria struct {
	TenantId        string `json:"tenantId"`
	ConsumerCode    string `json:"consumerCode"`
	BusinessService string `json:"businessService"`
}

type BillAccountDetail struct {
	Id                string        `json:"id,omitempty"`
	TenantId          string        `json:"tenantId,omitempty"`
	BillDetailId      string        `json:"billDetailId,omitempty"`
	DemandDetailId    string        `json:"demandDetailId,omitempty"`
	Order             int           `json:"order,omitempty"`
	Amount            float64       `json:"amount"`
	AdjustedAmount    float64       `json:"adjustedAmount"`
	TaxHeadCode       string        `json:"taxHeadCode"`
	AdditionalDetails interface{}   `json:"additionalDetails,omitempty"`
	AuditDetails      *AuditDetails `json:"auditDetails,omitempty"`
}

type BillDetail struct {
	Id                 string              `json:"id,omitempty"`
	TenantId           string              `json:"tenantId,omitempty"`
	DemandId           string              `json:"demandId,omitempty"`
	BillId             string              `json:"billId,omitempty"`
	ExpiryDate         int64               `json:"expiryDate,omitempty"`
	Amount             float64             `json:"amount"`
	FromPeriod         int64               `json:"fromPeriod,omitempty"`
	ToPeriod           int64               `json:"toPeriod,omitempty"`
	AdditionalDetails  interface{}         `json:"additionalDetails,omitempty"`
	BillAccountDetails []BillAccountDetail `json:"billAccountDetails,omitempty"`
}

type Bill struct {
	Id                string        `json:"id,omitempty"`
	MobileNumber      string        `json:"mobileNumber,omitempty"`
	PayerName         string        `json:"payerName,omitempty"`
	PayerAddress      string        `json:"payerAddress,omitempty"`
	PayerEmail        string        `json:"payerEmail,omitempty"`
	Status            string        `json:"status,omitempty"`
	TotalAmount       float64       `json:"totalAmount"`
	BusinessService   string        `json:"businessService,omitempty"`
	BillNumber        string        `json:"billNumber,omitempty"`
	BillDate          int64         `json:"billDate,omitempty"`
	ConsumerCode      string        `json:"consumerCode,omitempty"`
	AdditionalDetails interface{}   `json:"additionalDetails,omitempty"`
	BillDetails       []BillDetail  `json:"billDetails,omitempty"`
	TenantId          string        `json:"tenantId,omitempty"`
	AuditDetails      *AuditDetails `json:"auditDetails,omitempty"`
}

type BillResponse struct {
	ResponseInfo *ResponseInfo `json:"responseInfo,omitempty"`
	Bill         []Bill        `json:"bill"`
}

type Error struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Description string   `json:"description,omitempty"`
	Params      []string `json:"params,omitempty"`
}

type ErrorRes struct {
	ResponseInfo *ResponseInfo `json:"ResponseInfo,omitempty"`
	Errors       []Error       `json:"Errors"`
}

