package core

type CosmosAddress string

type CosmosNetwork string

const (
	CosmosHub CosmosNetwork = "cosmos-hub"
)

func (n CosmosNetwork) String() string {
	return string(n)
}
