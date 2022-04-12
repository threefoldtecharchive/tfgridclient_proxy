package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// to use for database/sql
	_ "github.com/lib/pq"
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
		COALESCE(sum(contract_resources.mru), 0) + 2147483648 as used_mru,
		COALESCE(sum(contract_resources.hru), 0) as used_hru,
		COALESCE(sum(contract_resources.sru), 0) as used_sru,
		node_resources_total.mru - COALESCE(sum(contract_resources.mru), 0) - 2147483648 as free_mru,
		node_resources_total.hru - COALESCE(sum(contract_resources.hru), 0) as free_hru,
		2 * node_resources_total.sru - COALESCE(sum(contract_resources.sru), 0) as free_sru,
		COALESCE(node_resources_total.cru, 0) as total_cru,
		COALESCE(node_resources_total.mru, 0) as total_mru,
		COALESCE(node_resources_total.hru, 0) as total_hru,
		COALESCE(node_resources_total.sru, 0) as total_sru
	FROM contract_resources
	JOIN node_contract as node_contract
	ON node_contract.id = contract_resources.contract_id
	RIGHT JOIN node as node
	ON node.node_id = node_contract.node_id
	JOIN node_resources_total AS node_resources_total
	ON node_resources_total.node_id = node.id
	WHERE node_contract.state = 'Created' OR node_contract.state IS NULL 
	GROUP BY node.node_id, node_resources_total.mru, node_resources_total.sru, node_resources_total.hru, node_resources_total.cru;

	CREATE OR REPLACE function node_resources(query_node_id INTEGER)
	returns table (node_id INTEGER, used_cru NUMERIC, used_mru NUMERIC, used_hru NUMERIC, used_sru NUMERIC, free_mru NUMERIC, free_hru NUMERIC, free_sru NUMERIC, total_cru NUMERIC, total_mru NUMERIC, total_hru NUMERIC, total_sru NUMERIC)
	as
	$body$
	SELECT
		node.node_id,
		COALESCE(sum(contract_resources.cru), 0) as used_cru,
		COALESCE(sum(contract_resources.mru), 0) + 2147483648 as used_mru,
		COALESCE(sum(contract_resources.hru), 0) as used_hru,
		COALESCE(sum(contract_resources.sru), 0) as used_sru,
		node_resources_total.mru - COALESCE(sum(contract_resources.mru), 0) - 2147483648 as free_mru,
		node_resources_total.hru - COALESCE(sum(contract_resources.hru), 0) as free_hru,
		2 * node_resources_total.sru - COALESCE(sum(contract_resources.sru), 0) as free_sru,
		COALESCE(node_resources_total.cru, 0) as total_cru,
		COALESCE(node_resources_total.mru, 0) as total_mru,
		COALESCE(node_resources_total.hru, 0) as total_hru,
		COALESCE(node_resources_total.sru, 0) as total_sru
	FROM contract_resources
	JOIN node_contract as node_contract
	ON node_contract.id = contract_resources.contract_id
	RIGHT JOIN node as node
	ON node.node_id = node_contract.node_id
	JOIN node_resources_total AS node_resources_total
	ON node_resources_total.node_id = node.id
	WHERE (node_contract.state = 'Created' OR node_contract.state IS NULL) AND node.node_id = query_node_id
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
		COALESCE(stellar_address, ''),
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
		COALESCE(node.certification_type, '')
	FROM node
	LEFT JOIN node_resources($1) ON node.node_id = node_resources.node_id
	LEFT JOIN public_config ON node.id = public_config.node_id
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
		COALESCE(node.certification_type, '')
	FROM node
	LEFT JOIN nodes_resources_view ON node.node_id = nodes_resources_view.node_id
	LEFT JOIN public_config ON node.id = public_config.node_id
	`
	selectFarmsWithFilter = `
	SELECT 
		farm_id,
		COALESCE(name, ''),
		COALESCE(twin_id, 0),
		COALESCE(pricing_policy_id, 0),
		COALESCE(stellar_address, ''),
		(
			SELECT 
				COALESCE(json_agg(json_build_object('id', id, 'ip', ip, 'contractId', contract_id, 'gateway', gateway)), '[]')
			FROM
				public_ip
			WHERE farm.id = public_ip.farm_id
		) as public_ips
	FROM farm
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
	(SELECT count(id) AS farms FROM farm),
	(SELECT count(DISTINCT country) AS countries FROM node)
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
func (d *PostgresDatabase) CountNodes() (int, error) {
	var count int
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
func (d *PostgresDatabase) GetCounters(filter StatsFilter) (Counters, error) {
	var counters Counters
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

func (d *PostgresDatabase) scanNode(rows *sql.Rows, node *AllNodeData) error {
	err := rows.Scan(
		&node.NodeData.ID,
		&node.NodeID,
		&node.NodeData.FarmID,
		&node.NodeData.TwinID,
		&node.NodeData.Country,
		&node.NodeData.GridVersion,
		&node.NodeData.City,
		&node.NodeData.Uptime,
		&node.NodeData.Created,
		&node.NodeData.FarmingPolicyID,
		&node.NodeData.UpdatedAt,
		&node.NodeData.TotalResources.CRU,
		&node.NodeData.TotalResources.SRU,
		&node.NodeData.TotalResources.HRU,
		&node.NodeData.TotalResources.MRU,
		&node.NodeData.UsedResources.CRU,
		&node.NodeData.UsedResources.SRU,
		&node.NodeData.UsedResources.HRU,
		&node.NodeData.UsedResources.MRU,
		&node.NodeData.PublicConfig.Domain,
		&node.NodeData.PublicConfig.Gw4,
		&node.NodeData.PublicConfig.Gw6,
		&node.NodeData.PublicConfig.Ipv4,
		&node.NodeData.PublicConfig.Ipv6,
		&node.NodeData.CertificationType,
	)
	if err != nil {
		return err
	}
	if int64(node.NodeData.UpdatedAt) >= time.Now().Unix()-nodeStateFactor*int64(reportInterval/time.Second) {
		node.NodeData.Status = "up"
	} else {
		node.NodeData.Status = "down"
	}
	return nil
}

func (d *PostgresDatabase) scanFarm(rows *sql.Rows, farm *Farm) error {
	var publicIPStr string
	err := rows.Scan(
		&farm.FarmID,
		&farm.Name,
		&farm.TwinID,
		&farm.PricingPolicyID,
		&farm.StellarAddress,
		&publicIPStr,
	)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(publicIPStr), &farm.PublicIps); err != nil {
		return err
	}
	return nil
}

// GetNode returns node info
func (d *PostgresDatabase) GetNode(nodeID uint32) (AllNodeData, error) {
	var node AllNodeData
	rows, err := d.db.Query(selectSingleNode, nodeID)
	if err != nil {
		return node, err
	}
	defer rows.Close()
	if !rows.Next() {
		return node, ErrNodeNotFound
	}
	err = d.scanNode(rows, &node)
	return node, err
}

// GetFarm return farm info
func (d *PostgresDatabase) GetFarm(farmID uint32) (Farm, error) {
	var farm Farm
	rows, err := d.db.Query(selectFarm, farmID)
	if err != nil {
		return farm, err
	}
	defer rows.Close()
	if !rows.Next() {
		return farm, ErrFarmNotFound
	}
	err = d.scanFarm(rows, &farm)
	return farm, err
}

func requiresFarmJoin(filter NodeFilter) bool {
	return filter.FarmName != nil || filter.FreeIPs != nil
}

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
func printQuery(query string, args ...interface{}) {
	for i, e := range args {
		query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i+1), convertParam(e))
	}
	fmt.Printf("node query: %s", query)
}

// GetNodes returns nodes filtered and paginated
func (d *PostgresDatabase) GetNodes(filter NodeFilter, limit Limit) ([]AllNodeData, error) {
	query := selectNodesWithFilter
	args := make([]interface{}, 0)
	if requiresFarmJoin(filter) {
		query = fmt.Sprintf("%s JOIN farm ON node.farm_id = farm.farm_id", query)
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
	query = fmt.Sprintf("%s ORDER BY node.node_id", query)
	query = fmt.Sprintf("%s LIMIT $%d OFFSET $%d;", query, idx, idx+1)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query nodes")
	}
	defer rows.Close()
	nodes := make([]AllNodeData, 0)
	for rows.Next() {
		var node AllNodeData
		if err := d.scanNode(rows, &node); err != nil {
			log.Error().Err(err).Msg("failed to scan returned node from database")
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetFarms return farms filtered and paginated
func (d *PostgresDatabase) GetFarms(filter FarmFilter, limit Limit) ([]Farm, error) {
	query := selectFarmsWithFilter
	query = fmt.Sprintf("%s WHERE TRUE", query)
	args := make([]interface{}, 0)
	idx := 1
	if filter.FreeIPs != nil {
		query = fmt.Sprintf("%s AND (SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id and public_ip.contract_id = 0) >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeIPs)
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
	query = fmt.Sprintf("%s ORDER BY farm.farm_id", query)
	query = fmt.Sprintf("%s LIMIT $%d OFFSET $%d;", query, idx, idx+1)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't query farms")
	}
	defer rows.Close()
	farms := make([]Farm, 0)
	for rows.Next() {
		var farm Farm
		if err := d.scanFarm(rows, &farm); err != nil {
			log.Error().Err(err).Msg("failed to scan returned farm from database")
			continue
		}
		farms = append(farms, farm)
	}
	return farms, nil
}
