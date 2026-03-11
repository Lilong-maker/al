package service

import (
	__ "al/proto"
	"al/srv/dasic/config"
	"al/srv/handler/model"
	"context"
	"fmt"
	"time"
)

// Server 实现 OrderServiceServer
type Server struct {
	__.UnimplementedOrderServiceServer
}

// 响应码常量
const (
	CodeSuccess = 0
	CodeError   = 1
)

// ============ Create (创建) ============
func (s *Server) OrderCreate(ctx context.Context, in *__.OrderCreateReq) (*__.OrderCreateResp, error) {
	// 检查必填字段
	if in.Name == "" {
		return &__.OrderCreateResp{
			Msg:  "订单名称不能为空",
			Code: CodeError,
		}, nil
	}

	// 创建订单模型
	order := &model.Order{
		Name:    in.Name,
		Price:   in.Price,
		Num:     int(in.Num),
		OrderSn: int(in.OrderSn),
	}

	// 保存到数据库
	if err := config.DB.Create(order).Error; err != nil {
		return &__.OrderCreateResp{
			Msg:  fmt.Sprintf("创建订单失败: %v", err),
			Code: CodeError,
		}, nil
	}

	return &__.OrderCreateResp{
		Msg:  "订单创建成功",
		Code: CodeSuccess,
		Data: modelToOrderInfo(order),
	}, nil
}

// ============ Read (查询) ============
func (s *Server) OrderGet(ctx context.Context, in *__.OrderGetReq) (*__.OrderGetResp, error) {
	// 检查 ID
	if in.Id <= 0 {
		return &__.OrderGetResp{
			Msg:  "订单 ID 无效",
			Code: CodeError,
		}, nil
	}

	// 查询订单
	var order model.Order
	if err := config.DB.First(&order, in.Id).Error; err != nil {
		return &__.OrderGetResp{
			Msg:  fmt.Sprintf("订单不存在: %v", err),
			Code: CodeError,
		}, nil
	}

	return &__.OrderGetResp{
		Msg:  "查询成功",
		Code: CodeSuccess,
		Data: modelToOrderInfo(&order),
	}, nil
}

// ============ List (三条件组合查询) ============
// 支持条件：
// 1. 名称模糊查询
// 2. 价格区间查询
// 3. 创建时间范围查询
func (s *Server) OrderList(ctx context.Context, in *__.OrderListReq) (*__.OrderListResp, error) {
	// 设置默认分页参数
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大分页大小
	}

	// ========== 构建三条件组合查询 ==========
	db := config.DB.Model(&model.Order{})
	var conditions []string // 用于记录查询条件

	// 【条件1】名称模糊查询（右模糊，可用索引）
	if in.Name != "" {
		db = db.Where("name LIKE ?", in.Name+"%")
		conditions = append(conditions, fmt.Sprintf("名称 LIKE '%s%%'", in.Name))
	}

	// 【条件2】价格区间查询
	if in.MinPrice > 0 {
		db = db.Where("price >= ?", in.MinPrice)
		conditions = append(conditions, fmt.Sprintf("价格 >= %.2f", in.MinPrice))
	}
	if in.MaxPrice > 0 {
		db = db.Where("price <= ?", in.MaxPrice)
		conditions = append(conditions, fmt.Sprintf("价格 <= %.2f", in.MaxPrice))
	}

	// 【条件3】创建时间范围查询
	if in.StartTime != "" {
		db = db.Where("created_at >= ?", in.StartTime)
		conditions = append(conditions, fmt.Sprintf("创建时间 >= %s", in.StartTime))
	}
	if in.EndTime != "" {
		db = db.Where("created_at <= ?", in.EndTime)
		conditions = append(conditions, fmt.Sprintf("创建时间 <= %s", in.EndTime))
	}

	// 打印查询条件（用于调试）
	fmt.Printf("[三条件查询] 条件: %v, 分页: page=%d, pageSize=%d\n",
		conditions, page, pageSize)

	// ========== 查询总数 ==========
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return &__.OrderListResp{
			Msg:  fmt.Sprintf("查询总数失败: %v", err),
			Code: CodeError,
		}, nil
	}

	// ========== 分页查询 ==========
	var orders []model.Order
	offset := (page - 1) * pageSize

	// 按 ID 降序排列，分页查询
	if err := db.Order("id DESC").
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&orders).Error; err != nil {
		return &__.OrderListResp{
			Msg:  fmt.Sprintf("查询订单列表失败: %v", err),
			Code: CodeError,
		}, nil
	}

	// 转换为响应数据
	orderInfos := make([]*__.OrderInfo, 0, len(orders))
	for _, order := range orders {
		orderInfos = append(orderInfos, modelToOrderInfo(&order))
	}

	return &__.OrderListResp{
		Msg:   "查询成功",
		Code:  CodeSuccess,
		Data:  orderInfos,
		Total: total,
	}, nil
}

// ============ Update (更新) ============
func (s *Server) OrderUpdate(ctx context.Context, in *__.OrderUpdateReq) (*__.OrderUpdateResp, error) {
	// 检查 ID
	if in.Id <= 0 {
		return &__.OrderUpdateResp{
			Msg:  "订单 ID 无效",
			Code: CodeError,
		}, nil
	}

	// 查询订单是否存在
	var order model.Order
	if err := config.DB.First(&order, in.Id).Error; err != nil {
		return &__.OrderUpdateResp{
			Msg:  fmt.Sprintf("订单不存在: %v", err),
			Code: CodeError,
		}, nil
	}

	// 更新字段
	updates := make(map[string]interface{})
	if in.Name != "" {
		updates["name"] = in.Name
	}
	if in.Price > 0 {
		updates["price"] = in.Price
	}
	if in.Num > 0 {
		updates["num"] = in.Num
	}
	if in.OrderSn > 0 {
		updates["order_sn"] = in.OrderSn
	}

	// 执行更新
	if err := config.DB.Model(&order).Updates(updates).Error; err != nil {
		return &__.OrderUpdateResp{
			Msg:  fmt.Sprintf("更新订单失败: %v", err),
			Code: CodeError,
		}, nil
	}

	// 重新查询获取最新数据
	config.DB.First(&order, in.Id)

	return &__.OrderUpdateResp{
		Msg:  "订单更新成功",
		Code: CodeSuccess,
		Data: modelToOrderInfo(&order),
	}, nil
}

// ============ Delete (删除) ============
func (s *Server) OrderDelete(ctx context.Context, in *__.OrderDeleteReq) (*__.OrderDeleteResp, error) {
	// 检查 ID
	if in.Id <= 0 {
		return &__.OrderDeleteResp{
			Msg:  "订单 ID 无效",
			Code: CodeError,
		}, nil
	}

	// 查询订单是否存在
	var order model.Order
	if err := config.DB.First(&order, in.Id).Error; err != nil {
		return &__.OrderDeleteResp{
			Msg:  fmt.Sprintf("订单不存在: %v", err),
			Code: CodeError,
		}, nil
	}

	// 删除订单（软删除）
	if err := config.DB.Delete(&order).Error; err != nil {
		return &__.OrderDeleteResp{
			Msg:  fmt.Sprintf("删除订单失败: %v", err),
			Code: CodeError,
		}, nil
	}

	return &__.OrderDeleteResp{
		Msg:  "订单删除成功",
		Code: CodeSuccess,
	}, nil
}

// ============ 辅助函数 ============
// modelToOrderInfo 将 model.Order 转换为 proto.OrderInfo
func modelToOrderInfo(order *model.Order) *__.OrderInfo {
	return &__.OrderInfo{
		Id:        int64(order.ID),
		Name:      order.Name,
		Price:     order.Price,
		Num:       int64(order.Num),
		OrderSn:   int64(order.OrderSn),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
		UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
	}
}

// ============ 查询优化建议 ============
/*
【索引优化】
在数据库执行以下 SQL 创建索引：

-- 单列索引
CREATE INDEX idx_name ON orders(name);
CREATE INDEX idx_price ON orders(price);
CREATE INDEX idx_created_at ON orders(created_at);

-- 组合索引（最左前缀原则）
-- 如果常用查询是：name + price + created_at
CREATE INDEX idx_name_price_time ON orders(name, price, created_at);

-- 如果常用查询是：price + created_at
CREATE INDEX idx_price_time ON orders(price, created_at);

【查询优化技巧】
1. 右模糊查询 'xxx%' 可以用索引，'%xxx%' 不能用索引
2. 时间范围查询确保字段有索引
3. 分页查询避免深分页（OFFSET 过大），使用游标分页

【示例 SQL】
-- 三条件组合查询
SELECT * FROM orders
WHERE name LIKE 'iPhone%'
  AND price BETWEEN 1000 AND 5000
  AND created_at BETWEEN '2024-01-01' AND '2024-12-31'
ORDER BY id DESC
LIMIT 10 OFFSET 0;
*/
