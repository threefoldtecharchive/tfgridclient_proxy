package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	ErrNodeNotFound = errors.New("node not found")
	ErrFarmNotFound = errors.New("farm not found")
)

const (
	// TODO: should use free_* for resources instead?
	// TODO: is updatedAt in graphql sufficient or should another field
	//       representing last graphql sync must be added?
	setupSql = `
	CREATE TABLE IF NOT EXISTS nodes (
		version INTEGER,
		id TEXT,
		node_id INTEGER PRIMARY KEY,
		farm_id INTEGER,
		twin_id INTEGER,
		country TEXT,
		grid_version INTEGER,
		city TEXT,
		uptime INTEGER,
		created INTEGER,
		farming_policy_id INTEGER,
		updated_at TEXT,
		total_cru INTEGER,
		total_sru INTEGER,
		total_hru INTEGER,
		total_mru INTEGER,
		total_ipv4 INTEGER,
		used_cru INTEGER,
		used_sru INTEGER,
		used_hru INTEGER,
		used_mru INTEGER,
		used_ipv4 INTEGER,
		domain TEXT,
		gw4 TEXT,
		gw6 TEXT,
		ipv4 TEXT,
		ipv6 TEXT,
		status TEXT,
		certification_type TEXT,
		zos_version TEXT,
		hypervisor TEXT,
		proxy_updated_at INTEGER, /* epoch of last update inside the proxy */
		last_node_error TEXT, /* last error encountered when getting node info */
		last_fetch_attempt INTEGER, /* last time the node got contacted */
		retries INTEGER /* number of times an error happened when contacting the node since last successful attempt*/
	);

	CREATE TABLE IF NOT EXISTS farms (
		farm_id INTEGER PRIMARY KEY,
		twin_id INTEGER,
		version INTEGER,
		pricing_policy_id INTEGER,
		stellar_address TEXT,
		public_ips TEXT,
		public_ip_count INTEGER
	);
	`
	insertNodeFromGraphql = `
	INSERT INTO nodes (
		version,
		id,
		node_id,
		farm_id,
		twin_id,
		country,
		grid_version,
		city,
		uptime,
		created,
		farming_policy_id,
		updated_at,
		total_cru,
		total_sru,
		total_hru,
		total_mru,
		total_ipv4,
		used_cru,
		used_sru,
		used_hru,
		used_mru,
		used_ipv4,
		domain,
		gw4,
		gw6,
		ipv4,
		ipv6,
		status,
		certification_type,
		zos_version,
		hypervisor,
		proxy_updated_at,
		last_node_error,
		last_fetch_attempt,
		retries
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ?, ?, ?, ?, ?, "init", ?, "", "", 0, "", 0, 0)
	ON CONFLICT DO UPDATE SET
		version = ?,
		id = ?,
		node_id = ?,
		farm_id = ?,
		twin_id = ?,
		country = ?,
		grid_version = ?,
		city = ?,
		uptime = ?,
		created = ?,
		farming_policy_id = ?,
		updated_at = ?,
		domain = ?,
		gw4 = ?,
		gw6 = ?,
		ipv4 = ?,
		ipv6 = ?,
		certification_type = ?
	WHERE node_id = ?
	`
	updateNodeData = `
	UPDATE nodes SET
		total_cru = ?,
		total_sru = ?,
		total_hru = ?,
		total_mru = ?,
		total_ipv4 = ?,
		used_cru = ?,
		used_sru = ?,
		used_hru = ?,
		used_mru = ?,
		used_ipv4 = ?,
		status = ?,
		zos_version = ?,
		hypervisor = ?,
		proxy_updated_at = ?,
		last_fetch_attempt = ?,
		retries = 0,
		last_node_error = ""
	WHERE node_id = ?
	`
	insertFarmFromGraphql = `
	INSERT INTO farms (
		farm_id,
		twin_id,
		version,
		pricing_policy_id,
		stellar_address,
		public_ips,
		public_ip_count
	)
	VALUES(?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT DO UPDATE SET
		farm_id = ?,
		twin_id = ?,
		version = ?,
		pricing_policy_id = ?,
		stellar_address = ?,
		public_ips = ?,
		public_ip_count = ?
	WHERE farm_id = ?
	;
`
	updateNodeError = `
	UPDATE nodes
	SET retries = retries + 1,
		last_node_error = ?,
		status = ?,
		last_fetch_attempt = ?
	WHERE node_id = ?;
	;
	`

	selectNode = `
	SELECT *
	FROM nodes
	WHERE node_id = ?
	`
	selectFarm = `
	SELECT *
	FROM farms
	WHERE farm_id = ?
	`
	selectNodesWithFilter = `
	SELECT *
	FROM nodes
	WHERE true
	`
)

type SqliteDatabase struct {
	mutex sync.RWMutex
	db    *sql.DB
}

func NewSqliteDatabase(filename string) (Database, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize db")
	}
	res := SqliteDatabase{sync.RWMutex{}, db}
	if err := res.initialize(); err != nil {
		return nil, errors.Wrap(err, "failed to setup tables")
	}
	return &res, nil
}

func (d *SqliteDatabase) initialize() error {
	_, err := d.db.Exec(setupSql)
	return err
}

func (d *SqliteDatabase) Close() error {
	return d.db.Close()
}

func (d *SqliteDatabase) InsertOrUpdateNodeGraphqlData(nodeID uint32, nodeInfo GraphqlData) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	args := []interface{}{
		nodeInfo.Version,
		nodeInfo.ID,
		nodeID,
		nodeInfo.FarmID,
		nodeInfo.TwinID,
		nodeInfo.Country,
		nodeInfo.GridVersion,
		nodeInfo.City,
		nodeInfo.Uptime,
		nodeInfo.Created,
		nodeInfo.FarmingPolicyID,
		nodeInfo.UpdatedAt,
		nodeInfo.PublicConfig.Domain,
		nodeInfo.PublicConfig.Gw4,
		nodeInfo.PublicConfig.Gw6,
		nodeInfo.PublicConfig.Ipv4,
		nodeInfo.PublicConfig.Ipv6,
		nodeInfo.CertificationType,
	}
	args = append(args, args...)
	args = append(args, nodeID)
	_, err := d.db.Exec(insertNodeFromGraphql, args...)
	return err
}
func (d *SqliteDatabase) UpdateNodeData(nodeID uint32, nodeInfo NodeData) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_, err := d.db.Exec(updateNodeData,
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
		time.Now().Unix(),
		nodeID,
	)
	return err
}
func (d *SqliteDatabase) UpdateNodeError(nodeID uint32, fetchErr error) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_, err := d.db.Exec(updateNodeError,
		fetchErr.Error(),
		"down",
		time.Now().Unix(),
		nodeID,
	)
	return err
}

func (d *SqliteDatabase) UpdateFarm(farmInfo Farm) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	publicIPStr, err := json.Marshal(farmInfo.PublicIps)
	if err != nil {
		return err
	}
	args := []interface{}{
		farmInfo.FarmID,
		farmInfo.TwinID,
		farmInfo.Version,
		farmInfo.PricingPolicyID,
		farmInfo.StellarAddress,
		string(publicIPStr),
		len(farmInfo.PublicIps),
	}
	args = append(args, args...)
	args = append(args, farmInfo.FarmID)
	_, err = d.db.Exec(insertFarmFromGraphql,
		args...,
	)
	return err
}

/*
	TODO: move to appropriate place

*/

func scanNode(rows *sql.Rows, node *AllNodeData) error {
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

func scanFarm(rows *sql.Rows, farm *Farm) error {
	var publicIPStr string
	err := rows.Scan(
		&farm.FarmID,
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

func (d *SqliteDatabase) GetNode(nodeID uint32) (AllNodeData, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	var node AllNodeData
	rows, err := d.db.Query(selectNode, nodeID)
	if err != nil {
		return node, err
	}
	defer rows.Close()
	if !rows.Next() {
		return node, ErrNodeNotFound
	}
	err = scanNode(rows, &node)
	return node, err
}

func (d *SqliteDatabase) GetFarm(farmID uint32) (Farm, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	var farm Farm
	rows, err := d.db.Query(selectFarm, farmID)
	if err != nil {
		return farm, err
	}
	defer rows.Close()
	if !rows.Next() {
		return farm, ErrFarmNotFound
	}
	err = scanFarm(rows, &farm)
	return farm, err
}

func (d *SqliteDatabase) GetNodes(filter NodeFilter, limit Limit) ([]AllNodeData, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	query := selectNodesWithFilter
	args := make([]interface{}, 0)
	if filter.Status != nil {
		query = fmt.Sprintf("%s AND status = ?", query)
		args = append(args, *filter.Status)
	}
	if filter.FreeCRU != nil {
		query = fmt.Sprintf("%s AND total_cru - used_cru >= ?", query)
		args = append(args, *filter.FreeCRU)
	}
	if filter.FreeMRU != nil {
		query = fmt.Sprintf("%s AND total_mru - used_mru >= ?", query)
		args = append(args, *filter.FreeMRU)
	}
	if filter.FreeHRU != nil {
		query = fmt.Sprintf("%s AND total_hru - used_hru >= ?", query)
		args = append(args, *filter.FreeHRU)
	}
	if filter.FreeSRU != nil {
		query = fmt.Sprintf("%s AND total_sru - used_sru >= ?", query)
		args = append(args, *filter.FreeSRU)
	}
	query = fmt.Sprintf("%s LIMIT ? OFFSET ?;", query)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		// TODO: revise error wrapping
		return nil, err
	}
	defer rows.Close()
	nodes := make([]AllNodeData, 0)
	for rows.Next() {
		var node AllNodeData
		if err := scanNode(rows, &node); err != nil {
			log.Error().Err(err).Msg("failed to scan returned node from database")
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (d *SqliteDatabase) GetFarms(filter FarmFilter, limit Limit) ([]Farm, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	query := selectNodesWithFilter
	args := make([]interface{}, 0)
	if filter.PublicIPCount != nil {
		query = fmt.Sprintf("%s AND public_ip_count >= ?", query)
		args = append(args, *filter.PublicIPCount)
	}
	// Q: most of these returns a single farm
	if filter.StellarAddress != nil {
		query = fmt.Sprintf("%s AND stellar_address = ?", query)
		args = append(args, *filter.StellarAddress)
	}
	if filter.PricingPolicyID != nil {
		query = fmt.Sprintf("%s AND pricing_policy_id = ?", query)
		args = append(args, *filter.PricingPolicyID)
	}
	if filter.Version != nil {
		query = fmt.Sprintf("%s AND version = ?", query)
		args = append(args, *filter.Version)
	}
	if filter.FarmID != nil {
		query = fmt.Sprintf("%s AND farm_id = ?", query)
		args = append(args, *filter.FarmID)
	}
	if filter.TwinID != nil {
		query = fmt.Sprintf("%s AND twin_id = ?", query)
		args = append(args, *filter.TwinID)
	}
	if filter.Name != nil {
		query = fmt.Sprintf("%s AND name = ?", query)
		args = append(args, *filter.Name)
	}
	query = fmt.Sprintf("%s LIMIT ? OFFSET ?;", query)
	args = append(args, limit.Size, (limit.Page-1)*limit.Size)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		// TODO: revise error wrapping
		return nil, err
	}
	defer rows.Close()
	farms := make([]Farm, 0)
	for rows.Next() {
		var farm Farm
		if err := scanFarm(rows, &farm); err != nil {
			log.Error().Err(err).Msg("failed to scan returned farm from database")
			continue
		}
		farms = append(farms, farm)
	}
	return farms, nil
}
