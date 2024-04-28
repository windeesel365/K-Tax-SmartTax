package sharedvars

import "database/sql"

// initialize value
var InitialPersonalExemption float64 = 60000.0
var Initialdonations float64 = 0.0
var InitialkReceipts float64 = 0.0

// initial exemptions กับค่า limits
var PersonalExemptionUpperLimit float64 = 100000.0
var DonationsUpperLimit float64 = 100000.0
var KReceiptsUpperLimit float64 = 50000.0

// declare สำหรับ ref database และ idข้อมูล postgresql
var Db *sql.DB
var Id int
