package query

import (
	"strings"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
)

func BuildSearchQuery(criteria domain.SearchCriteria) (string, []interface{}) {
	q := ChallanSearchQuery
	var args []interface{}
	args = append(args, criteria.TenantId)

	if criteria.ChallanNo != "" {
		ids := strings.Split(criteria.ChallanNo, ",")
		q += " AND e.challanno IN (?)"
		args = append(args, ids)
	}
	if criteria.Ids != "" {
		ids := strings.Split(criteria.Ids, ",")
		q += " AND e.id IN (?)"
		args = append(args, ids)
	}
	if criteria.AccountId != "" {
		q += " AND e.accountid = ?"
		args = append(args, criteria.AccountId)
	}
	if criteria.BusinessService != "" {
		q += " AND e.businessservice = ?"
		args = append(args, criteria.BusinessService)
	}
	if criteria.Status != "" {
		q += " AND e.applicationstatus = ?"
		args = append(args, criteria.Status)
	}
	if criteria.ReceiptNumber != "" {
		q += " AND e.receiptnumber = ?"
		args = append(args, criteria.ReceiptNumber)
	}
	if criteria.MobileNumber != "" {
		// e.accountid refers to uuid of citizen
		q += " AND e.accountid = ?"
		args = append(args, criteria.MobileNumber)
	}

	return q, args
}

const (
	ChallanSearchQuery = `SELECT e.id, e.tenantid, e.businessservice, e.challanno, e.referenceid, e.description, e.accountid, e.additionaldetail, e.source, e.taxperiodfrom, e.taxperiodto, e.applicationstatus, a.doorno, a.buildingname, a.street, a.city, a.pincode, a.latitude, a.longitude FROM eg_echallan e LEFT JOIN eg_challan_address a ON e.id = a.echallanid WHERE e.tenantid = ?`
	ChallanCountQuery  = `SELECT applicationstatus, COUNT(*) as count FROM eg_echallan WHERE tenantid = $1 GROUP BY applicationstatus`
)
