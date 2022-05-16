package main

// func unusedNecessaryColumns() string {
// 	return `
// 	INSERT INTO location VALUES ('1', '2', '3');
// 	`
// }
// func cleanup() string {
// 	return `
// 	DELETE FROM node_resources_total WHERE id = '1';
// 	DELETE FROM node WHERE id = '112';
// 	DELETE FROM location WHERE id = '1';
// 	DELETE FROM farm WHERE id = '112';
// 	UPDATE node_contract SET resources_used_id = NULL WHERE id = '112';
// 	DELETE FROM contract_resources WHERE id = '112';
// 	DELETE FROM node_contract WHERE id = '112';
// 	`
// }
// func test(db *sql.DB) error {
// 	node := node{
// 		id:          "112",
// 		location_id: "1",
// 		node_id:     112,
// 	}
// 	total_resources := node_resources_total{
// 		id:      "1",
// 		hru:     1,
// 		sru:     2,
// 		cru:     3,
// 		mru:     5,
// 		node_id: "112",
// 	}
// 	farm := farm{
// 		id:                 "112",
// 		farm_id:            112,
// 		name:               "345",
// 		certification_type: "123",
// 	}
// 	contract_resources := contract_resources{
// 		id:          "112",
// 		hru:         1,
// 		sru:         2,
// 		cru:         3,
// 		mru:         4,
// 		contract_id: "112",
// 	}
// 	node_contract := node_contract{
// 		id:                "112",
// 		contract_id:       112,
// 		twin_id:           112,
// 		node_id:           112,
// 		resources_used_id: "",
// 		deployment_data:   "123",
// 		deployment_hash:   "123",
// 		state:             "Created",
// 	}
// 	if _, err := db.Exec(unusedNecessaryColumns()); err != nil {
// 		return err
// 	}
// 	if _, err := db.Exec(insertQuery(&node)); err != nil {
// 		return err
// 	}
// 	if _, err := db.Exec(insertQuery(&total_resources)); err != nil {
// 		return err
// 	}
// 	if _, err := db.Exec(insertQuery(&farm)); err != nil {
// 		return err
// 	}
// 	if _, err := db.Exec(insertQuery(&node_contract)); err != nil {
// 		return err
// 	}
// 	if _, err := db.Exec(insertQuery(&contract_resources)); err != nil {
// 		return err
// 	}
// 	if _, err := db.Exec(setContractResource("112", "112")); err != nil {
// 		return err
// 	}
// 	return nil
// }
