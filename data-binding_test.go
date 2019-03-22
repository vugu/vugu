package vugu

// type Customer struct {
// 	Name string
// }

// type CustomerInfo struct {
// 	CustomerList []*Customer
// }

// type DemoCompData struct {
// 	DisplayCustomer *Customer
// }

// func init() {

// 	customerInfo := &CustomerInfo{
// 		CustomerList: []*Customer{
// 			&Customer{Name: "Joe"},
// 		},
// 	}

// 	demoCompData := &DemoCompData{}

// 	demoCompData.DisplayCustomer = customerInfo.CustomerList[0]

// 	// demoCompData needs to be notified of this change
// 	customerInfo.CustomerList[0].Name = "Bill"

// 	// hm, does demoCompData need to be notified of this???
// 	customerInfo.CustomerList = nil

// 	// if this were dynamic, it's like a "computed function" in Vue,
// 	// and keeps it fresh
// 	demoCompData.DisplayCustomer = func() *Customer {
// 		return customerInfo.CustomerList[0]
// 	}

// }
