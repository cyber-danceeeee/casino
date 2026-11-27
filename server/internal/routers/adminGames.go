package routers

import (
	"casino/internal/adminauth"
	"casino/internal/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminGamesRouter(group, handler.NewAdminGamesHandler()) })
}
func adminGamesRouter(group *gin.RouterGroup, h handler.AdminGamesHandler) {
	g := group.Group("/admin")
	g.Use(adminauth.Middleware())
	g.GET("/game-provider-connections", adminauth.Require(adminauth.PermissionGameProviderList), handler.ListProviderConnections)
	g.PUT("/game-provider-connections/:code", adminauth.Require(adminauth.PermissionGameProviderUpdate), handler.SaveProviderConnection)
	g.GET("/game-providers", adminauth.Require(adminauth.PermissionGameProviderList), h.ListProviders)
	g.GET("/game-providers/:code", adminauth.Require(adminauth.PermissionGameProviderView), h.GetProvider)
	g.PATCH("/game-providers/:code/status", adminauth.Require(adminauth.PermissionGameProviderUpdate), h.SetProviderStatus)
	g.PATCH("/game-providers/:code/maintenance", adminauth.Require(adminauth.PermissionGameProviderUpdate), h.SetProviderMaintenance)
	g.POST("/game-providers/:code/test", adminauth.Require(adminauth.PermissionGameProviderTest), h.TestProvider)
	g.POST("/game-providers/:code/sync", adminauth.Require(adminauth.PermissionGameProviderSync), h.SyncProvider)
	g.POST("/game-providers/:code/sync/cancel", adminauth.Require(adminauth.PermissionGameProviderSync), h.CancelProviderSync)
	g.DELETE("/game-providers/:code", adminauth.Require(adminauth.PermissionGameProviderUpdate), h.DeleteProvider)
	// Temporary compatibility route for admin clients deployed before provider deletion was renamed.
	g.DELETE("/game-providers/:code/games", adminauth.Require(adminauth.PermissionGameProviderUpdate), h.DeleteProvider)
	g.GET("/game-providers/:code/game-types", adminauth.Require(adminauth.PermissionGameList), h.ListProviderGameTypes)
	// Site game operations only need scoped type options. Keep the provider
	// management endpoint platform-only so API and callback configuration stays private.
	g.GET("/game-provider-options/:code/game-types", adminauth.Require(adminauth.PermissionGameList), h.ListProviderGameTypes)
	g.POST("/game-sync", adminauth.Require(adminauth.PermissionGameProviderSync), h.SyncAll)
	g.GET("/games", adminauth.Require(adminauth.PermissionGameList), h.ListGames)
	g.GET("/games/catalog", adminauth.Require(adminauth.PermissionGameList), h.ListImportableGames)
	g.GET("/games/catalog/provider-options/:code/game-types", adminauth.Require(adminauth.PermissionGameList), h.ListImportableProviderGameTypes)
	g.POST("/games/import", adminauth.Require(adminauth.PermissionGameUpdate), h.ImportGames)
	g.PATCH("/games/batch", adminauth.Require(adminauth.PermissionGameUpdate), h.BatchUpdateGames)
	g.GET("/games/:id", adminauth.Require(adminauth.PermissionGameView), h.GetGame)
	g.PATCH("/games/:id", adminauth.Require(adminauth.PermissionGameUpdate), h.UpdateGame)
	g.PATCH("/games/:id/status", adminauth.Require(adminauth.PermissionGameStatusUpdate), h.SetGameStatus)
	g.PUT("/games/:id/categories", adminauth.Require(adminauth.PermissionGameCategoryAssign), h.SetGameCategories)
	g.PUT("/games/:id/tags", adminauth.Require(adminauth.PermissionGameTagAssign), h.SetGameTags)
	g.PUT("/games/:id/collections", adminauth.Require(adminauth.PermissionGameCollectionAssign), h.SetGameCollections)
	g.POST("/game-category-mapping", adminauth.Require(adminauth.PermissionGameCategoryAssign), h.ApplyProviderCategoryMapping)
	g.POST("/games/:id/demo", adminauth.Require(adminauth.PermissionGameDemoLaunch), h.LaunchDemo)
	g.GET("/game-categories", adminauth.Require(adminauth.PermissionGameCategoryList), h.ListCategories)
	g.GET("/game-categories/:id", adminauth.Require(adminauth.PermissionGameCategoryList), h.GetCategory)
	g.GET("/game-categories/:id/games", adminauth.Require(adminauth.PermissionGameCategoryList), h.ListCategoryGames)
	g.GET("/game-categories/:id/candidate-games", adminauth.Require(adminauth.PermissionGameCategoryAssign), h.ListCategoryGameCandidates)
	g.GET("/game-categories/:id/operation-logs", adminauth.Require(adminauth.PermissionGameCategoryList), h.ListCategoryOperationLogs)
	g.POST("/game-categories", adminauth.Require(adminauth.PermissionGameCategoryCreate), h.CreateCategory)
	g.PATCH("/game-categories/:id", adminauth.Require(adminauth.PermissionGameCategoryUpdate), h.UpdateCategory)
	g.POST("/game-categories/:id/games", adminauth.Require(adminauth.PermissionGameCategoryAssign), h.AddCategoryGames)
	g.PATCH("/game-categories/:id/games/:gameID", adminauth.Require(adminauth.PermissionGameCategoryAssign), h.UpdateCategoryGameSort)
	g.DELETE("/game-categories/:id/games/:gameID", adminauth.Require(adminauth.PermissionGameCategoryAssign), h.RemoveCategoryGame)
	g.DELETE("/game-categories/:id", adminauth.Require(adminauth.PermissionGameCategoryDelete), h.DeleteCategory)
	g.GET("/game-tags", adminauth.Require(adminauth.PermissionGameTagList), h.ListTags)
	g.GET("/game-tags/:id", adminauth.Require(adminauth.PermissionGameTagList), h.GetTag)
	g.POST("/game-tags", adminauth.Require(adminauth.PermissionGameTagCreate), h.CreateTag)
	g.PATCH("/game-tags/:id", adminauth.Require(adminauth.PermissionGameTagUpdate), h.UpdateTag)
	g.PUT("/game-tags/:id/games", adminauth.Require(adminauth.PermissionGameTagAssign), h.SetTagGames)
	g.POST("/game-tags/:id/games", adminauth.Require(adminauth.PermissionGameTagAssign), h.SetTagGames)
	g.DELETE("/game-tags/:id", adminauth.Require(adminauth.PermissionGameTagDelete), h.DeleteTag)
	g.GET("/game-collections", adminauth.Require(adminauth.PermissionGameCollectionList), h.ListCollections)
	g.GET("/game-collections/:id", adminauth.Require(adminauth.PermissionGameCollectionList), h.GetCollection)
	g.POST("/game-collections", adminauth.Require(adminauth.PermissionGameCollectionCreate), h.CreateCollection)
	g.PATCH("/game-collections/:id", adminauth.Require(adminauth.PermissionGameCollectionUpdate), h.UpdateCollection)
	g.DELETE("/game-collections/:id", adminauth.Require(adminauth.PermissionGameCollectionDelete), h.DeleteCollection)
	g.GET("/game-sync-logs", adminauth.Require(adminauth.PermissionGameSyncLogView), h.ListSyncLogs)
	g.GET("/game-sync-logs/:id/items", adminauth.Require(adminauth.PermissionGameSyncLogView), h.ListSyncLogItems)
	g.GET("/game-sync-logs/:id/raw-response", adminauth.Require(adminauth.PermissionGameSyncLogView), h.GetSyncLogRawResponse)
	g.GET("/game-sync-logs/:id/operation-logs", adminauth.Require(adminauth.PermissionGameSyncLogView), h.ListSyncLogOperationLogs)
	g.GET("/game-sync-logs/:id", adminauth.Require(adminauth.PermissionGameSyncLogView), h.GetSyncLog)
	g.GET("/game-launch-logs", adminauth.Require(adminauth.PermissionGameLaunchLogView), h.ListLaunchLogs)
	g.GET("/game-launch-logs/:id", adminauth.Require(adminauth.PermissionGameLaunchLogView), h.GetLaunchLog)
	g.POST("/game-assets/upload", adminauth.Require(adminauth.PermissionGameAssetUpload), h.UploadGameAsset)
}
