package handler

import (
	"fmt"
	"strconv"
	"strings"
)

func h5IDOrderExpr(column string, ids []uint64) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(id, 10))
	}
	return fmt.Sprintf("FIELD(%s,%s)", column, strings.Join(parts, ","))
}
