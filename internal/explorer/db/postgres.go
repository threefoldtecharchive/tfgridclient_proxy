package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	ErrNodeNotFound = errors.New("node not found")
	ErrFarmNotFound = errors.New("farm not found")
)

const (
	setupPostgresql = `
	CREATE TABLE IF NOT EXISTS node_pulled (
		node_id INTEGER PRIMARY KEY,
		total_cru INTEGER,
		total_sru BIGINT,
		total_hru BIGINT,
		total_mru BIGINT,
		total_ipv4 INTEGER,
		used_cru INTEGER,
		used_sru BIGINT,
		used_hru BIGINT,
		used_mru BIGINT,
		used_ipv4 INTEGER,
		status TEXT,
		zos_version TEXT,
		hypervisor TEXT,
		proxy_updated_at BIGINT, /* epoch of last update inside the proxy */
		last_node_error TEXT, /* last error encountered when getting node info */
		last_fetch_attempt BIGINT, /* last time the node got contacted */
		retries INTEGER /* number of times an error happened when contacting the node since last successful attempt*/
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
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			'',
			$15,
			0
		)
	ON CONFLICT (node_id) DO UPDATE
	SET	total_cru = $2,
		total_sru = $3,
		total_hru = $4,
		total_mru = $5,
		total_ipv4 = $6,
		used_cru = $7,
		used_sru = $8,
		used_hru = $9,
		used_mru = $10,
		used_ipv4 = $11,
		status = $12,
		zos_version = $13,
		hypervisor = $14,
		proxy_updated_at = $15,
		last_fetch_attempt = $15,
		retries = 0,
		last_node_error = ''
	WHERE node_pulled.node_id = $1
	`
	updateNodeError = `
	INSERT INTO node_pulled
		VALUES (
			$1,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			$2,
			'',
			'',
			$3,
			$4,
			$3,
			1
		)
	ON CONFLICT (node_id) DO UPDATE
	SET retries = node_pulled.retries + 1,
		last_node_error = $4,
		status = $2,
		last_fetch_attempt = $3
	WHERE node_pulled.node_id = $1;
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
		COALESCE(node_pulled.total_cru, 0),
		COALESCE(node_pulled.total_sru, 0),
		COALESCE(node_pulled.total_hru, 0),
		COALESCE(node_pulled.total_mru, 0),
		COALESCE(node_pulled.used_ipv4, 0),
		COALESCE(node_pulled.used_cru, 0),
		COALESCE(node_pulled.used_sru, 0),
		COALESCE(node_pulled.used_hru, 0),
		COALESCE(node_pulled.used_mru, 0),
		COALESCE(node_pulled.used_ipv4, 0),
		COALESCE(node.public_config::json->'domain' #>> '{}', ''),
		COALESCE(node.public_config::json->'gw4' #>> '{}', ''),
		COALESCE(node.public_config::json->'gw6' #>> '{}', ''),
		COALESCE(node.public_config::json->'ipv4' #>> '{}', ''),
		COALESCE(node.public_config::json->'ipv6' #>> '{}', ''),
		COALESCE(node_pulled.status, 'init'),
		node.certification_type,
		COALESCE(node_pulled.zos_version, ''),
		COALESCE(node_pulled.hypervisor, ''),
		COALESCE(node_pulled.proxy_updated_at, 0),
		COALESCE(node_pulled.last_node_error, ''),
		COALESCE(node_pulled.last_fetch_attempt, 0),
		COALESCE(node_pulled.retries, 0)
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
)

type PostgresDatabase struct {
	db *sql.DB
}

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

func (d *PostgresDatabase) Close() error {
	return d.db.Close()
}
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
	err = rows.Scan(rows, &count)
	return count, err

}

func (d *PostgresDatabase) UpdateNodeData(nodeID uint32, nodeInfo NodeData) error {
	_, err := d.db.Exec(updateNodeData,
		nodeID,
		nodeInfo.TotalResources.CRU,
		nodeInfo.TotalResources.SRU,
		nodeInfo.TotalResources.MRU,
		nodeInfo.TotalResources.MRU,
		nodeInfo.TotalResources.IPV4U,
		nodeInfo.UsedResources.CRU,
		nodeInfo.UsedResources.SRU,
		nodeInfo.UsedResources.MRU,
		nodeInfo.UsedResources.MRU,
		nodeInfo.UsedResources.IPV4U,
		nodeInfo.Status,
		nodeInfo.ZosVersion,
		nodeInfo.Hypervisor,
		time.Now().Unix(),
	)
	return err
}
func (d *PostgresDatabase) UpdateNodeError(nodeID uint32, fetchErr error) error {
	_, err := d.db.Exec(updateNodeError,
		nodeID,
		"down",
		time.Now().Unix(),
		fetchErr.Error(),
	)
	return err
}

func (d *PostgresDatabase) scanNode(rows *sql.Rows, node *AllNodeData) error {
	return rows.Scan(
		&node.Graphql.Version,
		&node.Graphql.ID,
		&node.NodeID,
		&node.Graphql.FarmID,
		&node.Graphql.TwinID,
		&node.Graphql.Country,
		&node.Graphql.GridVersion,
		&node.Graphql.City,
		&node.Graphql.Uptime,
		&node.Graphql.Created,
		&node.Graphql.FarmingPolicyID,
		&node.Graphql.UpdatedAt,
		&node.Node.TotalResources.CRU,
		&node.Node.TotalResources.SRU,
		&node.Node.TotalResources.HRU,
		&node.Node.TotalResources.MRU,
		&node.Node.TotalResources.IPV4U,
		&node.Node.UsedResources.CRU,
		&node.Node.UsedResources.SRU,
		&node.Node.UsedResources.HRU,
		&node.Node.UsedResources.MRU,
		&node.Node.UsedResources.IPV4U,
		&node.Graphql.PublicConfig.Domain,
		&node.Graphql.PublicConfig.Gw4,
		&node.Graphql.PublicConfig.Gw6,
		&node.Graphql.PublicConfig.Ipv4,
		&node.Graphql.PublicConfig.Ipv6,
		&node.Node.Status,
		&node.Graphql.CertificationType,
		&node.Node.ZosVersion,
		&node.Node.Hypervisor,
		&node.ConnectionInfo.ProxyUpdateAt,
		&node.ConnectionInfo.LastNodeError,
		&node.ConnectionInfo.LastFetchAttempt,
		&node.ConnectionInfo.Retries,
	)
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

func (d *PostgresDatabase) GetNodes(filter NodeFilter, limit Limit) ([]AllNodeData, error) {
	query := selectNodesWithFilter
	args := make([]interface{}, 0)
	if requiresFarmJoin(filter) {
		query = fmt.Sprintf("%s JOIN farm ON node.farm_id = farm.farm_id", query)
	}
	idx := 1
	query = fmt.Sprintf("%s WHERE TRUE", query)
	if filter.Status != nil {
		query = fmt.Sprintf("%s AND node_pulled.status = $%d", query, idx)
		idx++
		args = append(args, *filter.Status)
	}
	if filter.FreeCRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.total_cru - node_pulled.used_cru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeCRU)
	}
	if filter.FreeMRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.total_mru - node_pulled.used_mru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeMRU)
	}
	if filter.FreeHRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.total_hru - node_pulled.used_hru >= $%d", query, idx)
		idx++
		args = append(args, *filter.FreeHRU)
	}
	if filter.FreeSRU != nil {
		query = fmt.Sprintf("%s AND node_pulled.total_sru - node_pulled.used_sru >= $%d", query, idx)
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
