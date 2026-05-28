package query

const (
	ChallanSearchQuery = `SELECT id, tenantid, businessservice, challanno, referenceid, description, accountid, source, taxperiodfrom, taxperiodto, applicationstatus FROM eg_echallan WHERE tenantid = ?`
	ChallanCountQuery  = `SELECT COUNT(*) FROM eg_echallan WHERE tenantid = $1`
)
