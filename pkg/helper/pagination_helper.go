package helper

import "zenith-on-stage/internal/model"

func BuildPaginationQuery(paginationReq *model.PaginationRequest) model.PaginationQuery {
	offset := (paginationReq.Page - 1) * paginationReq.Size

	orderClause := paginationReq.Sort
	if paginationReq.Order != "" {
		orderClause += " " + paginationReq.Order
	}

	return model.PaginationQuery{
		Offset:      offset,
		Limit:       paginationReq.Size,
		OrderClause: orderClause,
		SearchQuery: paginationReq.SearchQuery,
	}
}
