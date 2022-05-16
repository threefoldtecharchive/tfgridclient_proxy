package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// to use for database/sql
	_ "github.com/lib/pq"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	// ErrNodeNotFound node not found
	ErrNodeNotFound = errors.New("node not found")
	// ErrFarmNotFound farm not found
	ErrFarmNotFound = errors.New("farm not found")
)

const (
	nodeStateFactor = 3
	reportInterval  = time.Hour
	// the number of missed reports to mark the node down
	// if node reports every 5 mins, it's marked down if the last report is more than 15 mins in the past
)

const (
	setupPostgresql = `
	CREATE OR REPLACE VIEW nodes_resources_view AS SELECT
		node.node_id,
		COALESCE(sum(contract_resources.cru), 0) as used_cru,
		COALESCE(sum(contract_resources.mru), 0) + GREATEST(CAST((node_resources_total.mru / 10) AS bigint), 2147483648) as used_mru,
		COALESCE(sum(contract_resources.hru), 0) as used_hru,
		COALESCE(sum(contract_resources.sru), 0) + 107374182400 as used_sru,
		node_resources_total.mru - COALESCE(sum(contract_resources.mru), 0) - GREATEST(CAST((node_resources_total.mru / 10) AS bigint), 2147483648) as free_mru,
		node_resources_total.hru - COALESCE(sum(contract_resources.hru), 0) as free_hru,
		2 * node_resources_total.sru - COALESCE(sum(contract_resources.sru), 0) - 107374182400 as free_sru,
		COALESCE(node_resources_total.cru, 0) as total_cru,
		COALESCE(node_resources_total.mru, 0) as total_mru,
		COALESCE(node_resources_total.hru, 0) as total_hru,
		COALESCE(node_resources_total.sru, 0) as total_sru
	FROM contract_resources
	JOIN node_contract as node_contract
	ON node_contract.resources_used_id = contract_resources.id AND node_contract.state = 'Created'
	RIGHT JOIN node as node
	ON node.node_id = node_contract.node_id
	JOIN node_resources_total AS node_resources_total
	ON node_resources_total.node_id = node.id
	GROUP BY node.node_id, node_resources_total.mru, node_resources_total.sru, node_resources_total.hru, node_resources_total.cru;

	CREATE OR REPLACE function node_resources(query_node_id INTEGER)
	returns table (node_id INTEGER, used_cru NUMERIC, used_mru NUMERIC, used_hru NUMERIC, used_sru NUMERIC, free_mru NUMERIC, free_hru NUMERIC, free_sru NUMERIC, total_cru NUMERIC, total_mru NUMERIC, total_hru NUMERIC, total_sru NUMERIC)
	as
	$body$
	SELECT
		node.node_id,
		COALESCE(sum(contract_resources.cru), 0) as used_cru,
		COALESCE(sum(contract_resources.mru), 0) + GREATEST(CAST((node_resources_total.mru / 10) AS bigint), 2147483648) as used_mru,
		COALESCE(sum(contract_resources.hru), 0) as used_hru,
		COALESCE(sum(contract_resources.sru), 0) + 107374182400 as used_sru,
		node_resources_total.mru - COALESCE(sum(contract_resources.mru), 0) - GREATEST(CAST((node_resources_total.mru / 10) AS bigint), 2147483648) as free_mru,
		node_resources_total.hru - COALESCE(sum(contract_resources.hru), 0) as free_hru,
		2 * node_resources_total.sru - COALESCE(sum(contract_resources.sru), 0) - 107374182400 as free_sru,
		COALESCE(node_resources_total.cru, 0) as total_cru,
		COALESCE(node_resources_total.mru, 0) as total_mru,
		COALESCE(node_resources_total.hru, 0) as total_hru,
		COALESCE(node_resources_total.sru, 0) as total_sru
	FROM contract_resources
	JOIN node_contract as node_contract
	ON node_contract.resources_used_id = contract_resources.id AND node_contract.state = 'Created'
	RIGHT JOIN node as node
	ON node.node_id = node_contract.node_id
	JOIN node_resources_total AS node_resources_total
	ON node_resources_total.node_id = node.id
	WHERE node.node_id = query_node_id
	GROUP BY node.node_id, node_resources_total.mru, node_resources_total.sru, node_resources_total.hru, node_resources_total.cru;
	$body$
	language sql;
	`

	selectFarm = `
	SELECT 
		farm_id,
		COALESCE(name, ''),
		COALESCE(twin_id, 0),
		COALESCE(pricing_policy_id, 0),
		COALESCE(certification_type, ''),
		COALESCE(stellar_address, ''),
		COALESCE(dedicated_farm, false),
		(
			SELECT 
				COALESCE(json_agg(json_build_object('id', id, 'ip', ip, 'contractId', contract_id, 'gateway', gateway)), '[]')
			FROM
				public_ip
			WHERE farm.id = public_ip.farm_id
		) as public_ips
	FROM farm
	JOIN 
	WHERE farm.farm_id = $1
	`

	selectSingleNode = `
	SELECT
		node.id,
		COALESCE(node.node_id, 0),
		COALESCE(node.farm_id, 0),
		COALESCE(node.twin_id, 0),
		COALESCE(node.country, ''),
		COALESCE(node.grid_version, 0),
		COALESCE(node.city, ''),
		COALESCE(node.uptime, 0),
		COALESCE(node.created, 0),
		COALESCE(node.farming_policy_id, 0),
		COALESCE(CAST(updated_at / 1000 AS int), 0),
		COALESCE(node_resources.total_cru, 0),
		COALESCE(node_resources.total_sru, 0),
		COALESCE(node_resources.total_hru, 0),
		COALESCE(node_resources.total_mru, 0),
		COALESCE(node_resources.used_cru, 0),
		COALESCE(node_resources.used_sru, 0),
		COALESCE(node_resources.used_hru, 0),
		COALESCE(node_resources.used_mru, 0),
		COALESCE(public_config.domain, ''),
		COALESCE(public_config.gw4, ''),
		COALESCE(public_config.gw6, ''),
		COALESCE(public_config.ipv4, ''),
		COALESCE(public_config.ipv6, ''),
		COALESCE(node.certification_type, ''),
		COALESCE(farm.dedicated_farm, false),
		COALESCE(rent_contract.contract_id, 0),
		COALESCE(rent_contract.twin_id, 0),
		0
	FROM node
	LEFT JOIN node_resources($1) ON node.node_id = node_resources.node_id
	LEFT JOIN public_config ON node.id = public_config.node_id
	LEFT JOIN rent_contract ON rent_contract.state = 'Created' AND rent_contract.node_id = node.node_id
	LEFT JOIN farm ON node.farm_id = farm.farm_id
	WHERE node.node_id = $1;
	`
	selectNodesWithFilter = `
	SELECT
		node.id,
		COALESCE(node.node_id, 0),
		COALESCE(node.farm_id, 0),
		COALESCE(node.twin_id, 0),
		COALESCE(node.country, ''),
		COALESCE(node.grid_version, 0),
		COALESCE(node.city, ''),
		COALESCE(node.uptime, 0),
		COALESCE(node.created, 0),
		COALESCE(node.farming_policy_id, 0),
		COALESCE(CAST(updated_at / 1000 AS int), 0),
		COALESCE(nodes_resources_view.total_cru, 0),
		COALESCE(nodes_resources_view.total_sru, 0),
		COALESCE(nodes_resources_view.total_hru, 0),
		COALESCE(nodes_resources_view.total_mru, 0),
		COALESCE(nodes_resources_view.used_cru, 0),
		COALESCE(nodes_resources_view.used_sru, 0),
		COALESCE(nodes_resources_view.used_hru, 0),
		COALESCE(nodes_resources_view.used_mru, 0),
		COALESCE(public_config.domain, ''),
		COALESCE(public_config.gw4, ''),
		COALESCE(public_config.gw6, ''),
		COALESCE(public_config.ipv4, ''),
		COALESCE(public_config.ipv6, ''),
		COALESCE(node.certification_type, ''),
		COALESCE(farm.dedicated_farm, false),
		COALESCE(rent_contract.contract_id, 0),
		COALESCE(rent_contract.twin_id, 0),
		%s
	FROM node
	LEFT JOIN nodes_resources_view ON node.node_id = nodes_resources_view.node_id
	LEFT JOIN public_config ON node.id = public_config.node_id
	LEFT JOIN rent_contract ON rent_contract.state = 'Created' AND rent_contract.node_id = node.node_id
	LEFT JOIN farm ON node.farm_id = farm.farm_id
	`
	selectFarmsWithFilter = `
	SELECT 
		farm_id,
		COALESCE(name, ''),
		COALESCE(twin_id, 0),
		COALESCE(pricing_policy_id, 0),
		COALESCE(certification_type, ''),
		COALESCE(stellar_address, ''),
		COALESCE(dedicated_farm, false),
		(
			SELECT 
				COALESCE(json_agg(json_build_object('id', id, 'ip', ip, 'contractId', contract_id, 'gateway', gateway)), '[]')
			FROM
				public_ip
			WHERE farm.id = public_ip.farm_id
		) as public_ips,
		%s
	FROM farm
	`
	selectTwins = "SELECT twin_id, account_id, ip, %s From twin"

	selectContracts = `
	SELECT 
		contracts.contract_id,
	 	twin_id,
		state,
		CAST(created_at / 1000 AS int),
		name, 
		node_id, 
		deployment_data, 
		deployment_hash, 
		number_of_public_i_ps, 
		type,
		COALESCE(contract_billing.billings, '[]') as contract_billing, 
		%s 
	FROM (
	SELECT contract_id, twin_id, state, created_at, ''AS name, node_id, deployment_data, deployment_hash, number_of_public_i_ps, 'node' AS type
	FROM node_contract 
	UNION 
	SELECT contract_id, twin_id, state, created_at, '' AS name, node_id, '', '', 0, 'rent' AS type
	FROM rent_contract 
	UNION 
	SELECT contract_id, twin_id, state, created_at, name, 0, '', '', 0, 'name' AS type
	FROM name_contract
	) contracts
	LEFT JOIN (
		SELECT 
			contract_bill_report.contract_id,
			COALESCE(json_agg(json_build_object('amountBilled', amount_billed, 'discountReceived', discount_received, 'timestamp', timestamp)), '[]') as billings
		FROM
			contract_bill_report
		GROUP BY contract_id
	) contract_billing
	ON contracts.contract_id = contract_billing.contract_id
	`

	countNodes = `
	SELECT 
		count(*)
	FROM node
	`
	totalResources = `
	SELECT
		COALESCE(sum(node_resources_total.cru),0) AS total_cru,
		COALESCE(sum(node_resources_total.sru),0) AS total_sru,
		COALESCE(sum(node_resources_total.hru),0) AS total_hru,
		COALESCE(sum(node_resources_total.mru),0) AS total_mru
	FROM node
	LEFT JOIN node_resources_total ON node.id = node_resources_total.id
	`
	countersQuery = `
	SELECT
	(SELECT count(id) AS twins FROM twin),
	(SELECT count(id) AS public_ips FROM public_ip),
	(SELECT count(node.id) AS access_nodes FROM node 
		RIGHT JOIN public_config ON node.id = public_config.node_id 
		%[1]s AND (COALESCE(public_config.ipv4, '') != '' OR COALESCE(public_config.ipv4, '') != '')),
	(SELECT count(node.id) AS gateways FROM node 
	 	RIGHT JOIN public_config ON node.id = public_config.node_id 
	 	%[1]s AND COALESCE(public_config.domain, '') != '' AND (COALESCE(public_config.ipv4, '') != '' OR COALESCE(public_config.ipv6, '') != '')),
	(SELECT count(id) AS contracts FROM node_contract),
	(SELECT count(id) AS nodes FROM node %[1]s),
	(SELECT count(DISTINCT farm_id) AS farm FROM node %[1]s),
	(SELECT count(DISTINCT country) AS countries FROM node %[1]s)
	`
)

// PostgresDatabase postgres db client
type PostgresDatabase struct {
	db *sql.DB
}

// NewPostgresDatabase returns a new postgres db client
func NewPostgresDatabase(host string, port int, user, password, dbname string) (Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize db")
	}
	res := PostgresDatabase{db}
	if err := res.initialize(); err != nil {
		return nil, errors.Wrap(err, "failed to setup tables")
	}
	return &res, nil
}

// Close the db connection
func (d *PostgresDatabase) Close() error {
	return d.db.Close()
}

func (d *PostgresDatabase) initialize() error {
	_, err := d.db.Exec(setupPostgresql)
	return err
}

// CountNodes returns the total number of nodes
func (d *PostgresDatabase) CountNodes() (uint, error) {
	var count uint
	rows, err := d.db.Query(countNodes)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, errors.New("count query returned 0 rows")
	}
	err = rows.Scan(&count)
	return count, err

}

// GetCounters returns aggregate info about the grid
func (d *PostgresDatabase) GetCounters(filter types.StatsFilter) (types.Counters, error) {
	var counters types.Counters
	query := countersQuery
	totalResourcesQuery := totalResources

	if filter.Status != nil && *filter.Status == "up" {
		nodeUpInterval := time.Now().Unix()*1000 - nodeStateFactor*int64(reportInterval/time.Millisecond)
		condition := fmt.Sprintf(`WHERE node.updated_at >= %d`, nodeUpInterval)
		query = fmt.Sprintf(query, condition)
		totalResourcesQuery = fmt.Sprintf(`%s %s`, totalResourcesQuery, condition)
	} else {
		query = fmt.Sprintf(query, "where TRUE")
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return counters, errors.Wrap(err, "couldn't get counters")
	}
	defer rows.Close()
	if !rows.Next() {
		return counters, errors.New("count query returned 0 rows")
	}

	err = rows.Scan(
		&counters.Twins,
		&counters.PublicIPs,
		&counters.AccessNodes,
		&counters.Gateways,
		&counters.Contracts,
		&counters.Nodes,
		&counters.Farms,
		&counters.Countries,
	)
	if err != nil {
		return counters, errors.Wrap(err, "couldn't scan counters")
	}

	rows, err = d.db.Query(totalResourcesQuery)
	if err != nil {
		return counters, errors.Wrap(err, "couldn't query total resources")
	}
	defer rows.Close()
	if !rows.Next() {
		return counters, errors.New("total resources query returned 0 rows")
	}

	err = rows.Scan(
		&counters.TotalCRU,
		&counters.TotalSRU,
		&counters.TotalHRU,
		&counters.TotalMRU,
	)
	if err != nil {
		return counters, errors.Wrap(err, "couldn't scan total resources")
	}
	return counters, nil
}

func (d *PostgresDatabase) scanNode(rows *sql.Rows, node *DBNodeData, count *uint) error {
	err := rows.Scan(
		&node.ID,
		&node.NodeID,
		&node.FarmID,
		&node.TwinID,
		&node.Country,
		&node.GridVersion,
		&node.City,
		&node.Uptime,
		&node.Created,
		&node.FarmingPolicyID,
		&node.UpdatedAt,
		&node.TotalResources.CRU,
		&node.TotalResources.SRU,
		&node.TotalResources.HRU,
		&node.TotalResources.MRU,
		&node.UsedResources.CRU,
		&node.UsedResources.SRU,
		&node.UsedResources.HRU,
		&node.UsedResources.MRU,
		&node.PublicConfig.Domain,
		&node.PublicConfig.Gw4,
		&node.PublicConfig.Gw6,
		&node.PublicConfig.Ipv4,
		&node.PublicConfig.Ipv6,
		&node.CertificationType,
		&node.Dedicated,
		&node.RentContractID,
		&node.RentedByTwinID,
		count,
	)
	if err != nil {
		return err
	}
	if int64(node.UpdatedAt) >= time.Now().Unix()-nodeStateFactor*int64(reportInterval/time.Second) {
		node.Status = "up"
	} else {
		node.Status = "down"
	}
	return nil
}

func (d *PostgresDatabase) scanFarm(rows *sql.Rows, farm *types.Farm, count *uint) error {
	var publicIPStr string
	err := rows.Scan(
		&farm.FarmID,
		&farm.Name,
		&farm.TwinID,
		&farm.PricingPolicyID,
		&farm.CertificationType,
		&farm.StellarAddress,
		&farm.Dedicated,
		&publicIPStr,
		count,
	)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(publicIPStr), &farm.PublicIps); err != nil {
		return err
	}
	return nil
}

func (d *PostgresDatabase) scanTwin(rows *sql.Rows, twin *types.Twin, count *uint) error {
	err := rows.Scan(
		&twin.TwinID,
		&twin.AccountID,
		&twin.IP,
		count,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *PostgresDatabase) scanContract(rows *sql.Rows, contract *DBContract, count *uint) error {
	var contractBilling string
	err := rows.Scan(
		&contract.ContractID,
		&contract.TwinID,
		&contract.State,
		&contract.CreatedAt,
		&contract.Name,
		&contract.NodeID,
		&contract.DeploymentData,
		&contract.DeploymentHash,
		&contract.NumberOfPublicIps,
		&contract.Type,
		&contractBilling,
		count,
	)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(contractBilling), &contract.ContractBillings); err != nil {
		return err
	}
	return nil
}

// GetNode returns node info
func (d *PostgresDatabase) GetNode(nodeID uint32) (DBNodeData, error) {
	var node DBNodeData
	rows, err := d.db.Query(selectSingleNode, nodeID)
	if err != nil {
		return node, err
	}
	defer rows.Close()
	if !rows.Next() {
		return node, ErrNodeNotFound
	}
	var count uint
	err = d.scanNode(rows, &node, &count)
	return node, err
}

// GetFarm return farm info
func (d *PostgresDatabase) GetFarm(farmID uint32) (types.Farm, error) {
	var farm types.Farm
	rows, err := d.db.Query(selectFarm, farmID)
	if err != nil {
		return farm, err
	}
	defer rows.Close()
	if !rows.Next() {
		return farm, ErrFarmNotFound
	}
	var count uint
	err = d.scanFarm(rows, &farm, &count)
	return farm, err
}

//lint:ignore U1000 used for debugging
func convertParam(p interface{}) string {
	if v, ok := p.(string); ok {
		return fmt.Sprintf("'%s'", v)
	} else if v, ok := p.(uint64); ok {
		return fmt.Sprintf("%d", v)
	} else if v, ok := p.(int64); ok {
		return fmt.Sprintf("%d", v)
	} else if v, ok := p.(uint32); ok {
		return fmt.Sprintf("%d", v)
	} else if v, ok := p.(int); ok {
		return fmt.Sprintf("%d", v)
	} else if v, ok := p.(gridtypes.Unit); ok {
		return fmt.Sprintf("%d", v)
	}
	log.Error().Msgf("can't recognize type %s", fmt.Sprintf("%v", p))
	return "0"
}

//lint:ignore U1000 used for debugging
//nolint
func printQuery(query string, args ...interface{}) {
	for i, e := range args {
		query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i+1), convertParam(e))
	}
	fmt.Printf("node query: %s", query)
}

// GetNodes returns nodes filtered and paginated
func (d *PostgresDatabase) GetNodes(filter types.NodeFilter, limit types.Limit) ([]DBNodeData, uint, error) {
	query := selectNodesWithFilter
	args := make([]interface{}, 0)
	if limit.RetCount {
		query = fmt.Sprintf(query, "COUNT(*) OVER()")
	} else {
		query = fmt.Sprintf(query, "0")
	}
	idx := 1
	query = fmt.Sprintf("%s WHERE TRUE", query)
	if filter.Status != nil {
		if *filter.Status == "down" {
			query = fmt.Sprintf("%s AND (node.updated_at < $%d OR node.updated_at IS NULL)", query, idx)
		} else {
			query = fmt.Sprintf("%s AND node.updated_at >= $%d", query, idx)
		}
		idx++
		args = append(args, time.Now().Unix()*1000-nodeStateFactor*int64(reportInterval/time.Millisecond))
	}
	if filter.FreeMRU != nil {
		query = fmt.Sprintf("%s AND nodes_resources_view.free_mru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeMRU)
	}
	if filter.FreeHRU != nil {
		query = fmt.Sprintf("%s AND nodes_resources_view.free_hru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeHRU)
	}
	if filter.FreeSRU != nil {
		query = fmt.Sprintf("%s AND nodes_resources_view.free_sru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeSRU)
	}
	if filter.Country != nil {
		query = fmt.Sprintf("%s AND node.country = $%d", query, idx)
		idx++
		args = append(args, *filter.Country)
	}
	if filter.City != nil {
		query = fmt.Sprintf("%s AND node.city = $%d", query, idx)
		idx++
		args = append(args, *filter.City)
	}
	if filter.FarmIDs != nil {
		query = fmt.Sprintf("%s AND (false", query)
		for _, id := range filter.FarmIDs {
			query = fmt.Sprintf("%s OR node.farm_id = $%d", query, idx)
			idx++
			args = append(args, id)
		}
		query = fmt.Sprintf("%s)", query)
	}
	if filter.FarmName != nil {
		query = fmt.Sprintf("%s AND farm.name = $%d", query, idx)
		idx++
		args = append(args, *filter.FarmName)
	}
	if filter.FreeIPs != nil {
		query = fmt.Sprintf("%s AND (SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id AND public_ip.contract_id = 0) >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeIPs)
	}
	if filter.IPv4 != nil {
		query = fmt.Sprintf(`%s AND COALESCE(public_config.ipv4, '') != ''`, query)
	}
	if filter.IPv6 != nil {
		query = fmt.Sprintf(`%s AND COALESCE(public_config.ipv6, '') != ''`, query)
	}
	if filter.Domain != nil {
		query = fmt.Sprintf(`%s AND COALESCE(public_config.domain, '') != ''`, query)
	}
	if filter.Dedicated != nil {
		query = fmt.Sprintf(`%s AND farm.dedicated_farm = $%d`, query, idx)
		idx++
		args = append(args, *filter.Dedicated)
	}
	if filter.Rentable != nil {
		query = fmt.Sprintf(`%s AND ($%[2]d AND (farm.dedicated_farm = true AND COALESCE(rent_contract.contract_id, 0) = 0)
		OR NOT $%[2]d AND (farm.dedicated_farm = false OR (farm.dedicated_farm = true AND COALESCE(rent_contract.contract_id, 0) > 0)))`, query, idx)
		idx++
		args = append(args, *filter.Rentable)
	}
	if filter.RentedBy != nil {
		query = fmt.Sprintf("%s AND COALESCE(rent_contract.twin_id, 0) = $%d ", query, idx)
		idx++
		args = append(args, *filter.RentedBy)
	}
	if filter.AvailableFor != nil {
		query = fmt.Sprintf("%s AND (COALESCE(rent_contract.twin_id, 0) = $%d OR farm.dedicated_farm = false)", query, idx)
		idx++
		args = append(args, *filter.AvailableFor)
	}
	query = fmt.Sprintf("%s ORDER BY node.node_id", query)
	query = fmt.Sprintf("%s LIMIT $%d OFFSET $%d;", query, idx, idx+1)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to query nodes")
	}
	defer rows.Close()
	nodes := make([]DBNodeData, 0)
	var count uint
	for rows.Next() {
		var node DBNodeData
		if err := d.scanNode(rows, &node, &count); err != nil {
			log.Error().Err(err).Msg("failed to scan returned node from database")
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, count, nil
}

// GetFarms return farms filtered and paginated
func (d *PostgresDatabase) GetFarms(filter types.FarmFilter, limit types.Limit) ([]types.Farm, uint, error) {
	query := selectFarmsWithFilter
	if limit.RetCount {
		query = fmt.Sprintf(query, "COUNT(*) OVER()")
	} else {
		query = fmt.Sprintf(query, "0")
	}
	query = fmt.Sprintf("%s WHERE TRUE", query)
	args := make([]interface{}, 0)
	idx := 1
	if filter.FreeIPs != nil {
		query = fmt.Sprintf("%s AND (SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id and public_ip.contract_id = 0) >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeIPs)
	}
	if filter.TotalIPs != nil {
		query = fmt.Sprintf("%s AND (SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id) >= $%d", query, idx)
		idx++
		args = append(args, *filter.TotalIPs)
	}

	if filter.StellarAddress != nil {
		query = fmt.Sprintf("%s AND stellar_address = $%d", query, idx)
		idx++
		args = append(args, *filter.StellarAddress)
	}
	if filter.PricingPolicyID != nil {
		query = fmt.Sprintf("%s AND pricing_policy_id = $%d", query, idx)
		idx++
		args = append(args, *filter.PricingPolicyID)
	}
	if filter.Version != nil {
		query = fmt.Sprintf("%s AND version = $%d", query, idx)
		idx++
		args = append(args, *filter.Version)
	}
	if filter.FarmID != nil {
		query = fmt.Sprintf("%s AND farm_id = $%d", query, idx)
		idx++
		args = append(args, *filter.FarmID)
	}
	if filter.TwinID != nil {
		query = fmt.Sprintf("%s AND twin_id = $%d", query, idx)
		idx++
		args = append(args, *filter.TwinID)
	}
	if filter.Name != nil {
		query = fmt.Sprintf("%s AND name = $%d", query, idx)
		idx++
		args = append(args, *filter.Name)
	}
	if filter.NameContains != nil {
		query = fmt.Sprintf("%s AND name LIKE $%d", query, idx)
		idx++
		args = append(args, fmt.Sprintf("%[1]s%s%[1]s", "%", *filter.NameContains))
	}

	if filter.CertificationType != nil {
		query = fmt.Sprintf("%s AND certification_type = $%d", query, idx)
		idx++
		args = append(args, *filter.CertificationType)
	}

	if filter.Dedicated != nil {
		query = fmt.Sprintf("%s AND dedicated_farm = $%d", query, idx)
		idx++
		args = append(args, *filter.Dedicated)
	}
	query = fmt.Sprintf("%s ORDER BY farm.farm_id", query)
	query = fmt.Sprintf("%s LIMIT $%d OFFSET $%d;", query, idx, idx+1)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "couldn't query farms")
	}
	defer rows.Close()
	farms := make([]types.Farm, 0)
	var count uint
	for rows.Next() {
		var farm types.Farm
		if err := d.scanFarm(rows, &farm, &count); err != nil {
			log.Error().Err(err).Msg("failed to scan returned farm from database")
			continue
		}
		farms = append(farms, farm)
	}
	return farms, count, nil
}

// GetTwins returns twins filtered and paginated
func (d *PostgresDatabase) GetTwins(filter types.TwinFilter, limit types.Limit) ([]types.Twin, uint, error) {
	query := selectTwins
	args := make([]interface{}, 0)
	if limit.RetCount {
		query = fmt.Sprintf(query, "COUNT(*) OVER()")
	} else {
		query = fmt.Sprintf(query, "0")
	}
	idx := 1
	query = fmt.Sprintf("%s WHERE TRUE", query)
	if filter.TwinID != nil {
		query = fmt.Sprintf("%s AND twin_id = $%d", query, idx)
		idx++
		args = append(args, *filter.TwinID)
	}
	if filter.AccountID != nil {
		query = fmt.Sprintf("%s AND account_id = $%d", query, idx)
		idx++
		args = append(args, *filter.AccountID)
	}
	query = fmt.Sprintf("%s ORDER BY twin_id", query)
	query = fmt.Sprintf("%s LIMIT $%d OFFSET $%d;", query, idx, idx+1)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to query twins")
	}
	defer rows.Close()
	twins := make([]types.Twin, 0)
	var count uint
	for rows.Next() {
		var twin types.Twin
		if err := d.scanTwin(rows, &twin, &count); err != nil {
			log.Error().Err(err).Msg("failed to scan returned twin from database")
			continue
		}
		twins = append(twins, twin)
	}
	return twins, count, nil
}

// GetContracts returns contracts filtered and paginated
func (d *PostgresDatabase) GetContracts(filter types.ContractFilter, limit types.Limit) ([]DBContract, uint, error) {
	query := selectContracts
	args := make([]interface{}, 0)
	if limit.RetCount {
		query = fmt.Sprintf(query, "COUNT(*) OVER()")
	} else {
		query = fmt.Sprintf(query, "0")
	}
	idx := 1
	query = fmt.Sprintf("%s WHERE TRUE", query)
	if filter.Type != nil {
		query = fmt.Sprintf("%s AND type = $%d", query, idx)
		idx++
		args = append(args, *filter.Type)
	}
	if filter.State != nil {
		query = fmt.Sprintf("%s AND state = $%d", query, idx)
		idx++
		args = append(args, *filter.State)
	}
	if filter.TwinID != nil {
		query = fmt.Sprintf("%s AND twin_id = $%d", query, idx)
		idx++
		args = append(args, *filter.TwinID)
	}
	if filter.ContractID != nil {
		query = fmt.Sprintf("%s AND contracts.contract_id = $%d", query, idx)
		idx++
		args = append(args, *filter.ContractID)
	}
	if filter.NodeID != nil {
		query = fmt.Sprintf("%s AND node_id = $%d", query, idx)
		idx++
		args = append(args, *filter.NodeID)
	}
	if filter.NumberOfPublicIps != nil {
		query = fmt.Sprintf("%s AND number_of_public_i_ps >= $%d", query, idx)
		idx++
		args = append(args, *filter.NumberOfPublicIps)
	}
	if filter.Name != nil {
		query = fmt.Sprintf("%s AND name = $%d", query, idx)
		idx++
		args = append(args, *filter.Name)
	}
	if filter.DeploymentData != nil {
		query = fmt.Sprintf("%s AND deployment_data = $%d", query, idx)
		idx++
		args = append(args, *filter.DeploymentData)
	}
	if filter.DeploymentHash != nil {
		query = fmt.Sprintf("%s AND deployment_hash = $%d", query, idx)
		idx++
		args = append(args, *filter.DeploymentHash)
	}
	query = fmt.Sprintf("%s ORDER BY contracts.contract_id", query)
	query = fmt.Sprintf("%s LIMIT $%d OFFSET $%d;", query, idx, idx+1)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to query contracts")
	}
	defer rows.Close()
	contracts := make([]DBContract, 0)
	var count uint
	for rows.Next() {
		var contract DBContract
		if err := d.scanContract(rows, &contract, &count); err != nil {
			log.Error().Err(err).Msg("failed to scan returned contract from database")
			continue
		}
		contracts = append(contracts, contract)
	}
	return contracts, count, nil
}
