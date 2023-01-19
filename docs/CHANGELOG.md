# CHANGELOG

## v1.6.8 
- Introducing API V2 for both explorer/rmb endpoints.
- Removing `/gateways` endpoint and instead providing `is_gateway=true` query on `/nodes` endpoint.
- New object response for `/nodes` and `/nodes/1` that has nested capacity and nested location.
  ```go
  type Node2 struct {
	ID                string        
	NodeID            int           
	FarmID            int           
	TwinID            int           
	GridVersion       int           
	Uptime            int64         
	Created           int64         
	FarmingPolicyID   int           
	UpdatedAt         int64         
	Capacity          CapacityResult
	Location          Location      
	PublicConfig      PublicConfig  
	Status            string        
	CertificationType string        
	Dedicated         bool          
	RentContractID    uint          
	RentedByTwinID    uint          
	SerialNumber      string        
  }
  ```
- Visit `<grid-proxy-url>/api/v2/swagger/index.html` for detailed info.