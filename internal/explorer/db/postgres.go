package db

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	// to use for database/sql
	_ "github.com/lib/pq"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	// ErrNodeNotFound node not found
	ErrNodeNotFound = errors.New("node not found")
	// ErrFarmNotFound farm not found
	ErrFarmNotFound = errors.New("farm not found")
	//ErrViewNotFound
	ErrNodeResourcesViewNotFound = errors.New("ERROR: relation \"nodes_resources_view\" does not exist (SQLSTATE 42P01)")
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
		node_resources_total.sru - COALESCE(sum(contract_resources.sru), 0) - 107374182400 as free_sru,
		COALESCE(node_resources_total.cru, 0) as total_cru,
		COALESCE(node_resources_total.mru, 0) as total_mru,
		COALESCE(node_resources_total.hru, 0) as total_hru,
		COALESCE(node_resources_total.sru, 0) as total_sru,
		COALESCE(COUNT(DISTINCT state), 0) as states
	FROM contract_resources
	JOIN node_contract as node_contract
	ON node_contract.resources_used_id = contract_resources.id AND node_contract.state IN ('Created', 'GracePeriod')
	RIGHT JOIN node as node
	ON node.node_id = node_contract.node_id
	JOIN node_resources_total AS node_resources_total
	ON node_resources_total.node_id = node.id
	GROUP BY node.node_id, node_resources_total.mru, node_resources_total.sru, node_resources_total.hru, node_resources_total.cru;

	DROP FUNCTION IF EXISTS node_resources(query_node_id INTEGER);
	CREATE OR REPLACE function node_resources(query_node_id INTEGER)
	returns table (node_id INTEGER, used_cru NUMERIC, used_mru NUMERIC, used_hru NUMERIC, used_sru NUMERIC, free_mru NUMERIC, free_hru NUMERIC, free_sru NUMERIC, total_cru NUMERIC, total_mru NUMERIC, total_hru NUMERIC, total_sru NUMERIC, states BIGINT)
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
		node_resources_total.sru - COALESCE(sum(contract_resources.sru), 0) - 107374182400 as free_sru,
		COALESCE(node_resources_total.cru, 0) as total_cru,
		COALESCE(node_resources_total.mru, 0) as total_mru,
		COALESCE(node_resources_total.hru, 0) as total_hru,
		COALESCE(node_resources_total.sru, 0) as total_sru,
		COALESCE(COUNT(DISTINCT state), 0) as states
	FROM contract_resources
	JOIN node_contract as node_contract
	ON node_contract.resources_used_id = contract_resources.id AND node_contract.state IN ('Created', 'GracePeriod')
	RIGHT JOIN node as node
	ON node.node_id = node_contract.node_id
	JOIN node_resources_total AS node_resources_total
	ON node_resources_total.node_id = node.id
	WHERE node.node_id = query_node_id
	GROUP BY node.node_id, node_resources_total.mru, node_resources_total.sru, node_resources_total.hru, node_resources_total.cru;
	$body$
	language sql;

	DROP FUNCTION IF EXISTS convert_to_decimal(v_input text);
	CREATE OR REPLACE FUNCTION convert_to_decimal(v_input text)
	RETURNS DECIMAL AS $$
	DECLARE v_dec_value DECIMAL DEFAULT NULL;
	BEGIN
		BEGIN
			v_dec_value := v_input::DECIMAL;
		EXCEPTION WHEN OTHERS THEN
			RAISE NOTICE 'Invalid decimal value: "%".  Returning NULL.', v_input;
			RETURN NULL;
		END;
	RETURN v_dec_value;
	END;
	$$ LANGUAGE plpgsql;`
)

// PostgresDatabase postgres db client
type PostgresDatabase struct {
	gormDB *gorm.DB
}

// NewPostgresDatabase returns a new postgres db client
func NewPostgresDatabase(host string, port int, user, password, dbname string) (Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	gormDB, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create orm wrapper around db")
	}
	res := PostgresDatabase{gormDB}
	if err := res.initialize(); err != nil {
		return nil, errors.Wrap(err, "failed to setup tables")
	}
	return &res, nil
}

// Close the db connection
func (d *PostgresDatabase) Close() error {
	db, err := d.gormDB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *PostgresDatabase) initialize() error {
	res := d.gormDB.Exec(setupPostgresql)
	return res.Error
}

// GetCounters returns aggregate info about the grid
func (d *PostgresDatabase) GetCounters(filter types.StatsFilter) (types.Counters, error) {
	var counters types.Counters
	if res := d.gormDB.Table("twin").Count(&counters.Twins); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get twin count")
	}
	if res := d.gormDB.Table("public_ip").Count(&counters.PublicIPs); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get public ip count")
	}
	var count int64
	if res := d.gormDB.Table("node_contract").Count(&count); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get node contract count")
	}
	counters.Contracts += count
	if res := d.gormDB.Table("rent_contract").Count(&count); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get rent contract count")
	}
	counters.Contracts += count
	if res := d.gormDB.Table("name_contract").Count(&count); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get name contract count")
	}
	counters.Contracts += count
	if res := d.gormDB.Table("farm").Distinct("farm_id").Count(&counters.Farms); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get farm count")
	}

	condition := "TRUE"
	if filter.Status != nil {
		nodeUpInterval := time.Now().Unix() - nodeStateFactor*int64(reportInterval.Seconds())
		if *filter.Status == "up" {
			condition = fmt.Sprintf(`node.updated_at >= %d`, nodeUpInterval)
		} else if *filter.Status == "down" {
			condition = fmt.Sprintf(`node.updated_at < %d`, nodeUpInterval)
		}
	}

	if res := d.gormDB.
		Table("node").
		Select(
			"sum(node_resources_total.cru) as total_cru",
			"sum(node_resources_total.sru) as total_sru",
			"sum(node_resources_total.hru) as total_hru",
			"sum(node_resources_total.mru) as total_mru",
		).
		Joins("LEFT JOIN node_resources_total ON node.id = node_resources_total.node_id").
		Where(condition).
		Scan(&counters); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get nodes total resources")
	}
	if res := d.gormDB.Table("node").
		Where(condition).Count(&counters.Nodes); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get node count")
	}
	if res := d.gormDB.Table("node").
		Where(condition).Distinct("country").Count(&counters.Countries); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get country count")
	}
	query := d.gormDB.
		Table("node").
		Joins(
			`RIGHT JOIN public_config
			ON node.id = public_config.node_id
			`,
		)

	if res := query.Where(condition).Where("COALESCE(public_config.ipv4, '') != '' OR COALESCE(public_config.ipv6, '') != ''").Count(&counters.AccessNodes); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get access node count")
	}
	if res := query.Where(condition).Where("COALESCE(public_config.domain, '') != '' AND (COALESCE(public_config.ipv4, '') != '' OR COALESCE(public_config.ipv6, '') != '')").Count(&counters.Gateways); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get gateway count")
	}
	var distribution []NodesDistribution
	if res := d.gormDB.Table("node").
		Select("country, count(node_id) as nodes").Where(condition).Group("country").Scan(&distribution); res.Error != nil {
		return counters, errors.Wrap(res.Error, "couldn't get nodes distribution")
	}
	nodesDistribution := map[string]int64{}
	for _, d := range distribution {
		nodesDistribution[d.Country] = d.Nodes
	}
	counters.NodesDistribution = nodesDistribution
	return counters, nil
}

// GetNode returns node info
func (d *PostgresDatabase) GetNode(nodeID uint32) (Node, error) {
	q := d.nodeTableQuery()
	q = q.Where("node.node_id = ?", nodeID)
	q = q.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	var node Node
	res := q.Scan(&node)
	if d.shouldRetry(res.Error) {
		res = q.Scan(&node)
	}
	if res.Error != nil {
		return Node{}, res.Error
	}
	if node.ID == "" {
		return Node{}, ErrNodeNotFound
	}
	return node, nil
}

// GetFarm return farm info
func (d *PostgresDatabase) GetFarm(farmID uint32) (Farm, error) {
	q := d.farmTableQuery()
	q = q.Where("farm.farm_id = ?", farmID)
	var farm Farm
	if res := q.Scan(&farm); res.Error != nil {
		return farm, errors.Wrap(res.Error, "failed to scan returned farm from database")
	}
	return farm, nil
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

// nolint
//
//lint:ignore U1000 used for debugging
func printQuery(query string, args ...interface{}) {
	for i, e := range args {
		query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i+1), convertParam(e))
	}
	fmt.Printf("node query: %s", query)
}
func (d *PostgresDatabase) farmTableQuery() *gorm.DB {
	return d.gormDB.
		Table("farm").
		Select(
			"farm.farm_id",
			"name",
			"twin_id",
			"pricing_policy_id",
			"certification",
			"stellar_address",
			"dedicated_farm as dedicated",
			"COALESCE(public_ip.public_ips, '[]') as public_ips",
		).
		Joins(
			`LEFT JOIN
		(SELECT
			farm_id, 
			json_agg(json_build_object('id', id, 'ip', ip, 'contractId', contract_id, 'gateway', gateway)) as public_ips
		FROM
			public_ip
		GROUP by farm_id) public_ip
		ON public_ip.farm_id = farm.id`,
		)
}
func (d *PostgresDatabase) nodeTableQuery() *gorm.DB {
	return d.gormDB.
		Table("node").
		Select(
			"node.id",
			"node.node_id",
			"node.farm_id",
			"node.twin_id",
			"node.country",
			"node.grid_version",
			"node.city",
			"node.uptime",
			"node.created",
			"node.farming_policy_id",
			"updated_at",
			"nodes_resources_view.total_cru",
			"nodes_resources_view.total_sru",
			"nodes_resources_view.total_hru",
			"nodes_resources_view.total_mru",
			"nodes_resources_view.used_cru",
			"nodes_resources_view.used_sru",
			"nodes_resources_view.used_hru",
			"nodes_resources_view.used_mru",
			"public_config.domain",
			"public_config.gw4",
			"public_config.gw6",
			"public_config.ipv4",
			"public_config.ipv6",
			"node.certification",
			"farm.dedicated_farm as dedicated",
			"rent_contract.contract_id as rent_contract_id",
			"rent_contract.twin_id as rented_by_twin_id",
			"node.serial_number",
			"convert_to_decimal(location.longitude) as longitude",
			"convert_to_decimal(location.latitude) as latitude",
		).
		Joins(
			"LEFT JOIN nodes_resources_view ON node.node_id = nodes_resources_view.node_id",
		).
		Joins(
			"LEFT JOIN public_config ON node.id = public_config.node_id",
		).
		Joins(
			"LEFT JOIN rent_contract ON rent_contract.state IN ('Created', 'GracePeriod') AND rent_contract.node_id = node.node_id",
		).
		Joins(
			"LEFT JOIN farm ON node.farm_id = farm.farm_id",
		).
		Joins(
			"LEFT JOIN location ON node.location_id = location.id",
		)
}

// GetNodes returns nodes filtered and paginated
func (d *PostgresDatabase) GetNodes(filter types.NodeFilter, limit types.Limit) ([]Node, uint, error) {
	q := d.nodeTableQuery()
	q = q.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	if filter.Status != nil {
		// TODO: this shouldn't be in db
		threshold := time.Now().Unix() - nodeStateFactor*int64(reportInterval.Seconds())
		if *filter.Status == "down" {
			q = q.Where("node.updated_at < ? OR node.updated_at IS NULL", threshold)
		} else {
			q = q.Where("node.updated_at >= ?", threshold)
		}
	}
	if filter.FreeMRU != nil {
		q = q.Where("nodes_resources_view.free_mru >= ?", *filter.FreeMRU)
	}
	if filter.FreeHRU != nil {
		q = q.Where("nodes_resources_view.free_hru >= ?", *filter.FreeHRU)
	}
	if filter.FreeSRU != nil {
		q = q.Where("nodes_resources_view.free_sru >= ?", *filter.FreeSRU)
	}
	if filter.TotalCRU != nil {
		q = q.Where("nodes_resources_view.total_cru >= ?", *filter.TotalCRU)
	}
	if filter.TotalHRU != nil {
		q = q.Where("nodes_resources_view.total_hru >= ?", *filter.TotalHRU)
	}
	if filter.TotalMRU != nil {
		q = q.Where("nodes_resources_view.total_mru >= ?", *filter.TotalMRU)
	}
	if filter.TotalSRU != nil {
		q = q.Where("nodes_resources_view.total_sru >= ?", *filter.TotalSRU)
	}
	if filter.Country != nil {
		q = q.Where("LOWER(node.country) = LOWER(?)", *filter.Country)
	}
	if filter.CountryContains != nil {
		q = q.Where("node.country ILIKE '%' || ? || '%'", *filter.CountryContains)
	}
	if filter.City != nil {
		q = q.Where("LOWER(node.city) = LOWER(?)", *filter.City)
	}
	if filter.CityContains != nil {
		q = q.Where("node.city ILIKE '%' || ? || '%'", *filter.CityContains)
	}
	if filter.NodeID != nil {
		q = q.Where("node.node_id = ?", *filter.NodeID)
	}
	if filter.TwinID != nil {
		q = q.Where("node.twin_id = ?", *filter.TwinID)
	}
	if filter.FarmIDs != nil {
		q = q.Where("node.farm_id IN ?", filter.FarmIDs)
	}
	if filter.FarmName != nil {
		q = q.Where("LOWER(farm.name) = LOWER(?)", *filter.FarmName)
	}
	if filter.FarmNameContains != nil {
		q = q.Where("farm.name ILIKE '%' || ? || '%'", *filter.FarmNameContains)
	}
	if filter.FreeIPs != nil {
		q = q.Where("(SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id AND public_ip.contract_id = 0) >= ?", *filter.FreeIPs)
	}
	if filter.IPv4 != nil {
		q = q.Where("COALESCE(public_config.ipv4, '') != ''")
	}
	if filter.IPv6 != nil {
		q = q.Where("COALESCE(public_config.ipv6, '') != ''")
	}
	if filter.Domain != nil {
		q = q.Where("COALESCE(public_config.domain, '') != ''")
	}
	if filter.Dedicated != nil {
		q = q.Where("farm.dedicated_farm = ?", *filter.Dedicated)
	}
	if filter.Rentable != nil {
		q = q.Where(`? = ((farm.dedicated_farm = true OR nodes_resources_view.states = 0) AND COALESCE(rent_contract.contract_id, 0) = 0)`, *filter.Rentable)
	}
	if filter.RentedBy != nil {
		q = q.Where(`COALESCE(rent_contract.twin_id, 0) = ?`, *filter.RentedBy)
	}
	if filter.AvailableFor != nil {
		q = q.Where(`COALESCE(rent_contract.twin_id, 0) = ? OR (COALESCE(rent_contract.twin_id, 0) = 0 AND farm.dedicated_farm = false)`, *filter.AvailableFor)
	}
	if filter.Rented != nil {
		q = q.Where(`? = (COALESCE(rent_contract.contract_id, 0) != 0)`, *filter.Rented)
	}
	if filter.CertificationType != nil {
		q = q.Where("node.certification ILIKE ?", *filter.CertificationType)
	}

	var count int64
	if limit.Randomize || limit.RetCount {
		q = q.Session(&gorm.Session{})
		res := q.Count(&count)
		if d.shouldRetry(res.Error) {
			res = q.Count(&count)
		}
		if res.Error != nil {
			return nil, 0, res.Error
		}
	}
	if limit.Randomize {
		q = q.Limit(int(limit.Size)).
			Offset(int(rand.Intn(int(count)) - int(limit.Size)))
	} else {
		if filter.AvailableFor != nil {
			q = q.Order("(case when rent_contract is not null then 1 else 2 end)")
		}
		q = q.Limit(int(limit.Size)).
			Offset(int(limit.Page-1) * int(limit.Size)).
			Order("node_id")
	}

	var nodes []Node
	q = q.Session(&gorm.Session{})
	res := q.Scan(&nodes)
	if d.shouldRetry(res.Error) {
		res = q.Scan(&nodes)
	}
	if res.Error != nil {
		return nil, 0, res.Error
	}
	return nodes, uint(count), nil
}

func (d *PostgresDatabase) shouldRetry(resError error) bool {
	if resError != nil && resError.Error() == ErrNodeResourcesViewNotFound.Error() {
		if err := d.initialize(); err != nil {
			log.Logger.Err(err).Msg("failed to reinitialize database")
		} else {
			return true
		}
	}
	return false
}

// GetFarms return farms filtered and paginated
func (d *PostgresDatabase) GetFarms(filter types.FarmFilter, limit types.Limit) ([]Farm, uint, error) {
	q := d.farmTableQuery()
	if filter.FreeIPs != nil {
		q = q.Where("(SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id and public_ip.contract_id = 0) >= ?", *filter.FreeIPs)
	}
	if filter.TotalIPs != nil {
		q = q.Where("(SELECT count(id) from public_ip WHERE public_ip.farm_id = farm.id) >= ?", *filter.TotalIPs)
	}
	if filter.StellarAddress != nil {
		q = q.Where("stellar_address = ?", *filter.StellarAddress)
	}
	if filter.PricingPolicyID != nil {
		q = q.Where("pricing_policy_id = ?", *filter.PricingPolicyID)
	}
	if filter.FarmID != nil {
		q = q.Where("farm.farm_id = ?", *filter.FarmID)
	}
	if filter.TwinID != nil {
		q = q.Where("twin_id = ?", *filter.TwinID)
	}
	if filter.Name != nil {
		q = q.Where("LOWER(name) = LOWER(?)", *filter.Name)
	}

	if filter.NameContains != nil {
		escaped := strings.Replace(*filter.NameContains, "%", "\\%", -1)
		escaped = strings.Replace(escaped, "_", "\\_", -1)
		q = q.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", escaped))
	}

	if filter.CertificationType != nil {
		q = q.Where("certification = ?", *filter.CertificationType)
	}

	if filter.Dedicated != nil {
		q = q.Where("dedicated_farm = ?", *filter.Dedicated)
	}
	var count int64
	if limit.Randomize || limit.RetCount {
		if res := q.Count(&count); res.Error != nil {
			return nil, 0, errors.Wrap(res.Error, "couldn't get farm count")
		}
	}
	if limit.Randomize {
		q = q.Limit(int(limit.Size)).
			Offset(int(rand.Intn(int(count)) - int(limit.Size)))
	} else {
		q = q.Limit(int(limit.Size)).
			Offset(int(limit.Page-1) * int(limit.Size)).
			Order("farm.farm_id")
	}
	var farms []Farm
	if res := q.Scan(&farms); res.Error != nil {
		return farms, uint(count), errors.Wrap(res.Error, "failed to scan returned farm from database")
	}
	return farms, uint(count), nil
}

// GetTwins returns twins filtered and paginated
func (d *PostgresDatabase) GetTwins(filter types.TwinFilter, limit types.Limit) ([]types.Twin, uint, error) {
	q := d.gormDB.
		Table("twin").
		Select(
			"twin_id",
			"account_id",
			"relay",
			"public_key",
		)
	if filter.TwinID != nil {
		q = q.Where("twin_id = ?", *filter.TwinID)
	}
	if filter.AccountID != nil {
		q = q.Where("account_id = ?", *filter.AccountID)
	}
	if filter.Relay != nil {
		q = q.Where("relay = ?", *filter.Relay)
	}
	if filter.PublicKey != nil {
		q = q.Where("public_key = ?", *filter.PublicKey)
	}
	var count int64
	if limit.Randomize || limit.RetCount {
		if res := q.Count(&count); res.Error != nil {
			return nil, 0, errors.Wrap(res.Error, "couldn't get twin count")
		}
	}
	if limit.Randomize {
		q = q.Limit(int(limit.Size)).
			Offset(int(rand.Intn(int(count)) - int(limit.Size)))
	} else {
		q = q.Limit(int(limit.Size)).
			Offset(int(limit.Page-1) * int(limit.Size)).
			Order("twin.twin_id")
	}
	twins := []types.Twin{}

	if res := q.Scan(&twins); res.Error != nil {
		return twins, uint(count), errors.Wrap(res.Error, "failed to scan returned twins from database")
	}
	return twins, uint(count), nil
}

// GetContracts returns contracts filtered and paginated
func (d *PostgresDatabase) GetContracts(filter types.ContractFilter, limit types.Limit) ([]DBContract, uint, error) {
	q := d.gormDB.
		Table(`(SELECT contract_id, twin_id, state, created_at, ''AS name, node_id, deployment_data, deployment_hash, number_of_public_i_ps, 'node' AS type
	FROM node_contract 
	UNION 
	SELECT contract_id, twin_id, state, created_at, '' AS name, node_id, '', '', 0, 'rent' AS type
	FROM rent_contract 
	UNION 
	SELECT contract_id, twin_id, state, created_at, name, 0, '', '', 0, 'name' AS type
	FROM name_contract) contracts`).
		Select(
			"contracts.contract_id",
			"twin_id",
			"state",
			"created_at",
			"name",
			"node_id",
			"deployment_data",
			"deployment_hash",
			"number_of_public_i_ps as number_of_public_ips",
			"type",
			"COALESCE(contract_billing.billings, '[]') as contract_billings",
		).
		Joins(
			`LEFT JOIN (
				SELECT 
					contract_bill_report.contract_id,
					COALESCE(json_agg(json_build_object('amountBilled', amount_billed, 'discountReceived', discount_received, 'timestamp', timestamp)), '[]') as billings
				FROM
					contract_bill_report
				GROUP BY contract_id
			) contract_billing
			ON contracts.contract_id = contract_billing.contract_id`,
		)
	if filter.Type != nil {
		q = q.Where("type = ?", *filter.Type)
	}
	if filter.State != nil {
		q = q.Where("state ILIKE ?", *filter.State)
	}
	if filter.TwinID != nil {
		q = q.Where("twin_id = ?", *filter.TwinID)
	}
	if filter.ContractID != nil {
		q = q.Where("contracts.contract_id = ?", *filter.ContractID)
	}
	if filter.NodeID != nil {
		q = q.Where("node_id = ?", *filter.NodeID)
	}
	if filter.NumberOfPublicIps != nil {
		q = q.Where("number_of_public_i_ps >= ?", *filter.NumberOfPublicIps)
	}
	if filter.Name != nil {
		q = q.Where("name = ?", *filter.Name)
	}
	if filter.DeploymentData != nil {
		q = q.Where("deployment_data = ?", *filter.DeploymentData)
	}
	if filter.DeploymentHash != nil {
		q = q.Where("deployment_hash = ?", *filter.DeploymentHash)
	}
	var count int64
	if limit.Randomize || limit.RetCount {
		if res := q.Count(&count); res.Error != nil {
			return nil, 0, errors.Wrap(res.Error, "couldn't get contract count")
		}
	}
	if limit.Randomize {
		q = q.Limit(int(limit.Size)).
			Offset(int(rand.Intn(int(count)) - int(limit.Size)))
	} else {
		q = q.Limit(int(limit.Size)).
			Offset(int(limit.Page-1) * int(limit.Size)).
			Order("contract_id")
	}
	var contracts []DBContract
	if res := q.Scan(&contracts); res.Error != nil {
		return contracts, uint(count), errors.Wrap(res.Error, "failed to scan returned contracts from database")
	}
	return contracts, uint(count), nil
}
