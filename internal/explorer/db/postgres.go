package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// to use for database/sql
	_ "github.com/lib/pq"
	"github.com/threefoldtech/zos/cmds/modules/noded"
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
	nodeStateFactor = 2
	// the number of missed reports to mark the node down
	// if node reports every 5 mins, it's marked down if the last report is more than 15 mins in the past
)

const (
	setupPostgresql = `
	CREATE TABLE IF NOT EXISTS node_pulled (
		node_id INTEGER PRIMARY KEY,
		used_cru INTEGER,
		free_sru BIGINT,
		free_hru BIGINT,
		free_mru BIGINT,
		used_ipv4 INTEGER,
		zos_version TEXT,
		hypervisor TEXT,
		proxy_updated_at BIGINT /* epoch of last update inside the proxy */
	);
	`
	updateNodeData = `
	INSERT INTO node_pulled
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9
		)
	ON CONFLICT (node_id) DO UPDATE
	SET	used_cru = $2,
		free_sru = $3,
		free_hru = $4,
		free_mru = $5,
		used_ipv4 = $6,
		zos_version = $7,
		hypervisor = $8,
		proxy_updated_at = $9
	WHERE node_pulled.node_id = $1
	`
	updateNodeDataByTwin = `
	INSERT INTO node_pulled
		VALUES (
			(select node_id from node where twin_id = $1),
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9
		)
	ON CONFLICT (node_id) DO UPDATE
	SET	used_cru = $2,
		free_sru = $3,
		free_hru = $4,
		free_mru = $5,
		used_ipv4 = $6,
		zos_version = $7,
		hypervisor = $8,
		proxy_updated_at = $9
	WHERE node_pulled.node_id = (select node_id from node where twin_id = $1)
	`

	selectFarm = `
	SELECT 
		farm_id,
		COALESCE(name, ''),
		COALESCE(twin_id, 0),
		COALESCE(version, 0),
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
	selectNodesWithFilter = `
	SELECT
		node.version,
		node.id,
		node.node_id,
		node.farm_id,
		node.twin_id,
		node.country,
		node.grid_version,
		node.city,
		node.uptime,
		node.created,
		node.farming_policy_id,
		node.updated_at,
		COALESCE(node.cru, 0),
		COALESCE(node.sru, 0),
		COALESCE(node.hru, 0),
		COALESCE(node.mru, 0),
		COALESCE(node_pulled.used_cru, 0),
		COALESCE(node_pulled.free_sru, 0),
		COALESCE(node_pulled.free_hru, 0),
		COALESCE(node_pulled.free_mru, 0),
		COALESCE(node_pulled.used_ipv4, 0),
		COALESCE(node.public_config::json->'domain' #>> '{}', ''),
		COALESCE(node.public_config::json->'gw4' #>> '{}', ''),
		COALESCE(node.public_config::json->'gw6' #>> '{}', ''),
		COALESCE(node.public_config::json->'ipv4' #>> '{}', ''),
		COALESCE(node.public_config::json->'ipv6' #>> '{}', ''),
		node.certification_type,
		COALESCE(node_pulled.zos_version, ''),
		COALESCE(node_pulled.hypervisor, ''),
		COALESCE(node_pulled.proxy_updated_at, 0)
	FROM node
	LEFT JOIN node_pulled ON node.node_id = node_pulled.node_id
	`
	selectFarmsWithFilter = `
	SELECT 
		farm_id,
		COALESCE(name, ''),
		COALESCE(twin_id, 0),
		COALESCE(version, 0),
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
		sum(cru) AS total_cru,
		sum(sru) AS total_sru,
		sum(hru) AS total_hru,
		sum(mru) AS total_mru
	FROM node;
	`
	countersQuery = `
	SELECT
	(SELECT count(id) AS twins FROM twin),
	(SELECT count(id) AS public_ips FROM public_ip),
	(SELECT count(id) AS access_nodes FROM node where node.public_config::json->'ipv4' #>> '{}' != '' OR node.public_config::json->'ipv4' #>> '{}' != ''),
	(SELECT count(id) AS gateways FROM node where node.public_config::json->'domain' #>> '{}' != '' AND (node.public_config::json->'ipv4' #>> '{}' != '' OR node.public_config::json->'ipv6' #>> '{}' != '')),
	(SELECT count(id) AS contracts FROM node_contract),
	(SELECT count(id) AS nodes FROM node),
	(SELECT count(id) AS farms FROM farm),
	(SELECT count(DISTINCT country) AS countries FROM node);
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

func (d *PostgresDatabase) initialize() error {
	_, err := d.db.Exec(setupPostgresql)
	return err
}

// Close the db connection
func (d *PostgresDatabase) Close() error {
	return d.db.Close()
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
func (d *PostgresDatabase) GetCounters() (Counters, error) {
	var counters Counters
	rows, err := d.db.Query(countersQuery)
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
	rows, err = d.db.Query(totalResources)
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

// UpdateNodeData update the database by the pulled data from the node
func (d *PostgresDatabase) UpdateNodeData(nodeID uint32, nodeInfo PulledNodeData) error {
	_, err := d.db.Exec(updateNodeData,
		nodeID,
		nodeInfo.Resources.UsedCRU,
		nodeInfo.Resources.FreeSRU,
		nodeInfo.Resources.FreeMRU,
		nodeInfo.Resources.FreeMRU,
		nodeInfo.Resources.UsedIPV4U,
		nodeInfo.ZosVersion,
		nodeInfo.Hypervisor,
		time.Now().Unix(),
	)
	return err
}

// UpdateNodeDataByTwin update the database by the pulled data from the node given its twin id
func (d *PostgresDatabase) UpdateNodeDataByTwin(twinID uint32, nodeInfo PulledNodeData) error {
	_, err := d.db.Exec(updateNodeDataByTwin,
		twinID,
		nodeInfo.Resources.UsedCRU,
		nodeInfo.Resources.FreeSRU,
		nodeInfo.Resources.FreeMRU,
		nodeInfo.Resources.FreeMRU,
		nodeInfo.Resources.UsedIPV4U,
		nodeInfo.ZosVersion,
		nodeInfo.Hypervisor,
		time.Now().Unix(),
	)
	return err
}

func (d *PostgresDatabase) scanNode(rows *sql.Rows, node *AllNodeData) error {
	err := rows.Scan(
		&node.NodeData.Version,
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
		&node.PulledNodeData.Resources.UsedCRU,
		&node.PulledNodeData.Resources.FreeSRU,
		&node.PulledNodeData.Resources.FreeHRU,
		&node.PulledNodeData.Resources.FreeMRU,
		&node.PulledNodeData.Resources.UsedIPV4U,
		&node.NodeData.PublicConfig.Domain,
		&node.NodeData.PublicConfig.Gw4,
		&node.NodeData.PublicConfig.Gw6,
		&node.NodeData.PublicConfig.Ipv4,
		&node.NodeData.PublicConfig.Ipv6,
		&node.NodeData.CertificationType,
		&node.PulledNodeData.ZosVersion,
		&node.PulledNodeData.Hypervisor,
		&node.ProxyUpdatedAt,
	)
	if err != nil {
		return err
	}
	if int64(node.ProxyUpdatedAt) >= time.Now().Unix()-nodeStateFactor*int64(noded.ReportInterval/time.Second) {
		node.PulledNodeData.Status = "up"
	} else {
		node.PulledNodeData.Status = "down"
	}
	return nil
}

func (d *PostgresDatabase) scanFarm(rows *sql.Rows, farm *Farm) error {
	var publicIPStr string
	err := rows.Scan(
		&farm.FarmID,
		&farm.Name,
		&farm.TwinID,
		&farm.Version,
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
	query := fmt.Sprintf("%s WHERE node.node_id = $1", selectNodesWithFilter)
	rows, err := d.db.Query(query, nodeID)
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
		op := ">="
		if *filter.Status == "down" {
			op = "<"
		}
		query = fmt.Sprintf("%s AND node_pulled.proxy_updated_at %s $%d", query, op, idx)
		idx++
		args = append(args, time.Now().Unix()-nodeStateFactor*int64(noded.ReportInterval/time.Second))
	}
	if filter.FreeMRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.free_mru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeMRU)
	}
	if filter.FreeHRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.free_hru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeHRU)
	}
	if filter.FreeSRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.free_sru >= $%d", query, idx)
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
		query = fmt.Sprintf(`%s AND COALESCE(node.public_config::json->'ipv4' #>> '{}', '') != ''`, query)
	}
	if filter.IPv6 != nil {
		query = fmt.Sprintf(`%s AND COALESCE(node.public_config::json->'ipv6' #>> '{}', '') != ''`, query)
	}
	if filter.Domain != nil {
		query = fmt.Sprintf(`%s AND COALESCE(node.public_config::json->'domain' #>> '{}', '') != ''`, query)
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
