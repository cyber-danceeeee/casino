package handler

import (
	"casino/internal/database"
	"casino/internal/game/connections"
	"casino/internal/gameadmin"

	"github.com/gin-gonic/gin"
)

func ListProviderConnections(c *gin.Context) {
	data, err := gameadmin.New(database.GetDB()).ListProviderConnections()
	gameWrite(c, data, err)
}
func SaveProviderConnection(c *gin.Context) {
	var input connections.Input
	if !gameBind(c, &input) {
		return
	}
	input.ProviderCode = c.Param("code")
	data, err := gameadmin.New(database.GetDB()).SaveProviderConnection(input, gameActor(c))
	gameWrite(c, data, err)
}
