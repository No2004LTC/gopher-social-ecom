package dto

type AdminDashboardStatsResponse struct {
  TotalUsers       int64 `json:"total_users"`
  TotalPosts       int64 `json:"total_posts"`
  TotalBannedUsers int64 `json:"total_banned_users"`
}

// GrowthPoint đại diện cho số lượng mới trong 1 ngày
type GrowthPoint struct {
  Date       string `json:"date"`
  UsersCount int64  `json:"users_count"`
  PostsCount int64  `json:"posts_count"`
}
