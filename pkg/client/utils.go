package client

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/threefoldtech/grid_proxy_server/pkg/types"
)

func stringifyList(l []uint64) string {
	var ls []string
	for _, v := range l {
		ls = append(ls, strconv.FormatUint(v, 10))
	}
	return strings.Join(ls, ",")
}

func nodeParams(filter types.NodeFilter, limit types.Limit) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "?")

	if filter.Status != nil {
		fmt.Fprintf(&builder, "status=%s&", url.QueryEscape(*filter.Status))
	}
	if filter.FreeMRU != nil && *filter.FreeMRU != 0 {
		fmt.Fprintf(&builder, "free_mru=%d&", *filter.FreeMRU)
	}
	if filter.FreeHRU != nil && *filter.FreeHRU != 0 {
		fmt.Fprintf(&builder, "free_hru=%d&", *filter.FreeHRU)
	}
	if filter.FreeSRU != nil && *filter.FreeSRU != 0 {
		fmt.Fprintf(&builder, "free_sru=%d&", *filter.FreeSRU)
	}
	if filter.TotalCRU != nil && *filter.TotalCRU != 0 {
		fmt.Fprintf(&builder, "total_cru=%d&", *filter.TotalCRU)
	}
	if filter.TotalHRU != nil && *filter.TotalHRU != 0 {
		fmt.Fprintf(&builder, "total_hru=%d&", *filter.TotalHRU)
	}
	if filter.TotalMRU != nil && *filter.TotalMRU != 0 {
		fmt.Fprintf(&builder, "total_mru=%d&", *filter.TotalMRU)
	}
	if filter.TotalSRU != nil && *filter.TotalSRU != 0 {
		fmt.Fprintf(&builder, "total_sru=%d&", *filter.TotalSRU)
	}
	if filter.Country != nil && *filter.Country != "" {
		fmt.Fprintf(&builder, "country=%s&", url.QueryEscape(*filter.Country))
	}
	if filter.CountryContains != nil && *filter.CountryContains != "" {
		fmt.Fprintf(&builder, "country_contains=%s&", url.QueryEscape(*filter.CountryContains))
	}
	if filter.City != nil && *filter.City != "" {
		fmt.Fprintf(&builder, "city=%s&", url.QueryEscape(*filter.City))
	}
	if filter.CityContains != nil && *filter.CityContains != "" {
		fmt.Fprintf(&builder, "city_contains=%s&", url.QueryEscape(*filter.CityContains))
	}
	if filter.FarmName != nil && *filter.FarmName != "" {
		fmt.Fprintf(&builder, "farm_name=%s&", url.QueryEscape(*filter.FarmName))
	}
	if filter.FarmNameContains != nil && *filter.FarmNameContains != "" {
		fmt.Fprintf(&builder, "farm_name_contains=%s&", url.QueryEscape(*filter.FarmNameContains))
	}
	if filter.FarmIDs != nil && len(filter.FarmIDs) != 0 {
		fmt.Fprintf(&builder, "farm_ids=%s&", url.QueryEscape(stringifyList(filter.FarmIDs)))
	}
	if filter.FreeIPs != nil && *filter.FreeIPs != 0 {
		fmt.Fprintf(&builder, "free_ips=%d&", *filter.FreeIPs)
	}
	if filter.IPv4 != nil {
		fmt.Fprintf(&builder, "ipv4=%t&", *filter.IPv4)
	}
	if filter.IPv6 != nil {
		fmt.Fprintf(&builder, "ipv6=%t&", *filter.IPv6)
	}
	if filter.Domain != nil {
		fmt.Fprintf(&builder, "domain=%t&", *filter.Domain)
	}
	if filter.Rentable != nil {
		fmt.Fprintf(&builder, "rentable=%t&", *filter.Rentable)
	}
	if filter.NodeID != nil {
		fmt.Fprintf(&builder, "node_id=%d&", *filter.NodeID)
	}
	if filter.TwinID != nil {
		fmt.Fprintf(&builder, "twin_id=%d&", *filter.TwinID)
	}
	if filter.Rented != nil {
		fmt.Fprintf(&builder, "rented=%t&", *filter.Rented)
	}
	if filter.RentedBy != nil {
		// passing 0 might be helpful to get available non-rented nodes
		fmt.Fprintf(&builder, "rented_by=%d&", *filter.RentedBy)
	}
	if filter.AvailableFor != nil {
		fmt.Fprintf(&builder, "available_for=%d&", *filter.AvailableFor)
	}
	if limit.Page != 0 {
		fmt.Fprintf(&builder, "page=%d&", limit.Page)
	}
	if limit.Size != 0 {
		fmt.Fprintf(&builder, "size=%d&", limit.Size)
	}
	if limit.RetCount {
		fmt.Fprintf(&builder, "ret_count=true&")
	}
	if limit.Randomize {
		fmt.Fprintf(&builder, "randomize=true&")
	}
	if filter.CertificationType != nil && *filter.CertificationType != "" {
		fmt.Fprintf(&builder, "certification_type=%s&", url.QueryEscape(*filter.CertificationType))
	}

	res := builder.String()
	// pop the extra ? or &
	return res[:len(res)-1]
}

func farmParams(filter types.FarmFilter, limit types.Limit) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "?")

	if filter.FreeIPs != nil && *filter.FreeIPs != 0 {
		fmt.Fprintf(&builder, "free_ips=%d&", *filter.FreeIPs)
	}
	if filter.TotalIPs != nil && *filter.TotalIPs != 0 {
		fmt.Fprintf(&builder, "total_ips=%d&", *filter.TotalIPs)
	}
	if filter.StellarAddress != nil && *filter.StellarAddress != "" {
		fmt.Fprintf(&builder, "stellar_address=%s&", url.QueryEscape(*filter.StellarAddress))
	}
	if filter.PricingPolicyID != nil {
		fmt.Fprintf(&builder, "pricing_policy_id=%d&", *filter.PricingPolicyID)
	}
	if filter.FarmID != nil && *filter.FarmID != 0 {
		fmt.Fprintf(&builder, "farm_id=%d&", *filter.FarmID)
	}
	if filter.TwinID != nil && *filter.TwinID != 0 {
		fmt.Fprintf(&builder, "twin_id=%d&", *filter.TwinID)
	}
	if filter.Name != nil && *filter.Name != "" {
		fmt.Fprintf(&builder, "name=%s&", url.QueryEscape(*filter.Name))
	}
	if filter.NameContains != nil && *filter.NameContains != "" {
		fmt.Fprintf(&builder, "name_contains=%s&", url.QueryEscape(*filter.NameContains))
	}
	if filter.CertificationType != nil && *filter.CertificationType != "" {
		fmt.Fprintf(&builder, "certification_type=%s&", url.QueryEscape(*filter.CertificationType))
	}
	if filter.Dedicated != nil {
		fmt.Fprintf(&builder, "dedicated=%t&", *filter.Dedicated)
	}
	if limit.Page != 0 {
		fmt.Fprintf(&builder, "page=%d&", limit.Page)
	}
	if limit.Size != 0 {
		fmt.Fprintf(&builder, "size=%d&", limit.Size)
	}
	if limit.RetCount {
		fmt.Fprintf(&builder, "ret_count=true&")
	}
	if limit.Randomize {
		fmt.Fprintf(&builder, "randomize=true&")
	}

	res := builder.String()
	// pop the extra ? or &
	return res[:len(res)-1]
}

func twinParams(filter types.TwinFilter, limit types.Limit) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "?")

	if filter.TwinID != nil && *filter.TwinID != 0 {
		fmt.Fprintf(&builder, "twin_id=%d&", *filter.TwinID)
	}

	if filter.AccountID != nil && *filter.AccountID != "" {
		fmt.Fprintf(&builder, "account_id=%s&", url.QueryEscape(*filter.AccountID))
	}

	if limit.Page != 0 {
		fmt.Fprintf(&builder, "page=%d&", limit.Page)
	}
	if limit.Size != 0 {
		fmt.Fprintf(&builder, "size=%d&", limit.Size)
	}
	if limit.RetCount {
		fmt.Fprintf(&builder, "ret_count=true&")
	}
	if limit.Randomize {
		fmt.Fprintf(&builder, "randomize=true&")
	}

	res := builder.String()
	// pop the extra ? or &
	return res[:len(res)-1]
}

func contractParams(filter types.ContractFilter, limit types.Limit) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "?")

	if filter.ContractID != nil && *filter.ContractID != 0 {
		fmt.Fprintf(&builder, "contract_id=%d&", *filter.ContractID)
	}

	if filter.TwinID != nil && *filter.TwinID != 0 {
		fmt.Fprintf(&builder, "twin_id=%d&", *filter.TwinID)
	}
	if filter.NodeID != nil && *filter.NodeID != 0 {
		fmt.Fprintf(&builder, "node_id=%d&", *filter.NodeID)
	}
	if filter.Type != nil && *filter.Type != "" {
		fmt.Fprintf(&builder, "type=%s&", url.QueryEscape(*filter.Type))
	}
	if filter.State != nil && *filter.State != "" {
		fmt.Fprintf(&builder, "state=%s&", url.QueryEscape(*filter.State))
	}
	if filter.Name != nil && *filter.Name != "" {
		fmt.Fprintf(&builder, "name=%s&", url.QueryEscape(*filter.Name))
	}

	if filter.NumberOfPublicIps != nil && *filter.NumberOfPublicIps != 0 {
		fmt.Fprintf(&builder, "number_of_public_ips=%d&", *filter.NumberOfPublicIps)
	}
	if filter.DeploymentData != nil && *filter.DeploymentData != "" {
		fmt.Fprintf(&builder, "deployment_data=%s&", url.QueryEscape(*filter.DeploymentData))
	}
	if filter.DeploymentHash != nil && *filter.DeploymentHash != "" {
		fmt.Fprintf(&builder, "deployment_hash=%s&", url.QueryEscape(*filter.DeploymentHash))
	}
	if limit.Page != 0 {
		fmt.Fprintf(&builder, "page=%d&", limit.Page)
	}
	if limit.Size != 0 {
		fmt.Fprintf(&builder, "size=%d&", limit.Size)
	}
	if limit.RetCount {
		fmt.Fprintf(&builder, "ret_count=true&")
	}
	if limit.Randomize {
		fmt.Fprintf(&builder, "randomize=true&")
	}

	res := builder.String()
	// pop the extra ? or &
	return res[:len(res)-1]
}

func statsParams(filter types.StatsFilter) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "?")

	if filter.Status != nil && *filter.Status != "" {
		fmt.Fprintf(&builder, "status=%s&", url.QueryEscape(*filter.Status))
	}

	res := builder.String()
	// pop the extra ? or &
	return res[:len(res)-1]
}
