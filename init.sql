-- 创建库存日志表
CREATE TABLE IF NOT EXISTS stock_log (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键id',
  product_id BIGINT NOT NULL COMMENT '商品ID',
  product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
  type TINYINT NOT NULL COMMENT '操作类型(1:入库,2:出库,3:退货)',
  quantity INT NOT NULL COMMENT '操作数量',
  before_stock INT NOT NULL COMMENT '操作前库存',
  after_stock INT NOT NULL COMMENT '操作后库存',
  order_no VARCHAR(64) COMMENT '关联订单号(出库/退货时)',
  remark VARCHAR(255) COMMENT '备注',
  operator VARCHAR(255) NOT NULL COMMENT '操作人',
  operator_type TINYINT NOT NULL COMMENT '操作人类型(1:用户,2:系统,3:管理员)',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  INDEX idx_product_id (product_id),
  INDEX idx_type (type),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存日志表'; 