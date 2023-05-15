package migration

import contracts "github.com/estafette/estafette-ci-contracts"

type PagedBuildsResponse struct {
	Items      []*contracts.Build   `json:"items"`
	Pagination contracts.Pagination `json:"pagination"`
}
